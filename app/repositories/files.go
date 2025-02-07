package repositories_v1

import (
	"fmt"
	"goshaka/app/models"
	"goshaka/app/models/scopes"
	"goshaka/configs"
	"goshaka/database"
	"goshaka/helpers"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// Show all files belong to current user
//
//	param pagination helpers.Pagination
//	param uid string
//	return *helpers.Pagination
func FileShowAll(pagination helpers.Pagination, uid string) *helpers.Pagination {
	db := database.DB
	var files []*models.File

	db.Scopes(scopes.Paginate(files, &pagination, db)).Where("user_id = ?", uid).Find(&files)

	// activate AWS session
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String(configs.GetEnv("AWS_DEFAULT_REGION")),
		Credentials: credentials.NewStaticCredentials(configs.GetEnv("AWS_ACCESS_KEY_ID"), configs.GetEnv("AWS_SECRET_ACCESS_KEY"), ""),
	})
	if err != nil {
		panic("cannot get s3 session")
	}

	svc := s3.New(sess)

	// loop file
	for i := range files {
		req, _ := svc.GetObjectRequest(&s3.GetObjectInput{
			Bucket: aws.String(configs.GetEnv("AWS_BUCKET")),
			Key:    aws.String(files[i].Filename),
		})

		path, _ := req.Presign(15 * time.Minute)

		files[i].Path = path
	}

	pagination.Rows = files

	return &pagination
}

// Show detail file
//
//	param 	id	string
//	param 	uid	string
//	return	models.File
func FileShow(id, uid string) models.File {
	db := database.DB
	var file models.File

	db.Find(&file, "id = ? and user_id = ?", id, uid)

	path, _ := helpers.GetPresignAWSS3(file.Filename)
	file.Path = path

	return file
}

// Create new file record
//
//	param	arr interface{}
//	return	models.File, error
func FileCreate(arr interface{}) (models.File, error) {
	db := database.DB

	file := &models.File{}
	err := db.Model(file).Create(arr).Error
	var newFile models.File
	db.Last(&newFile)

	return newFile, err
}

// Destroy a user's file
//
//	param 	id	string
//	param 	uid	string
//	return	models.File, error
func FileDestroy(id, uid string) (models.File, error) {
	db := database.DB
	var file models.File

	db.Find(&file, "id = ? and user_id = ?", id, uid)

	if file.ID == 0 {
		return file, fmt.Errorf("not found")
	}

	// delete from S3
	go func() {
		helpers.DeleteFromAWSS3(file.Filename)
	}()

	// To soft delete, just remove .Unscoped()
	err := db.Unscoped().Delete(&file).Error

	return file, err
}
