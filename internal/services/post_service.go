package services

import (
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/db"
	"golang-mongodb-rest-api-starter/internal/models"
	"golang-mongodb-rest-api-starter/internal/models/builders"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.uber.org/zap"
)

type PostServiceWrapper interface {
	GetById(post *models.Post, id string) error
}

type PostService struct {
	Collection *mongo.Collection
	Log        *zap.SugaredLogger
}

func NewPostService(db *mongo.Client, dbName string) *PostService {
	return &PostService{
		Collection: db.Database(dbName).Collection(constants.PostCollection)}
}

func (s *PostService) GetById(post *models.Post, id, organizationId string) error {
	objectId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "organizationId", Value: organizationId}, {Key: "_id", Value: objectId}}
	return db.FindOne(s.Collection, filter, post, nil)
}

func (s *PostService) Create(req *models.CreatePostRequest, organizationId, requestUserId string) (*models.Post, error) {
	post := models.Post{}
	post.Name = req.Name
	post.CreatedAt = time.Now()
	post.OrganizationId = organizationId
	post.CreatedById = requestUserId
	res, err := db.InsertOne(s.Collection, &post)
	post.Id = res.InsertedID.(primitive.ObjectID).Hex()
	return &post, err
}

func (s *PostService) Update(req *models.UpdatePostRequest, oldPost *models.Post, requestUserId string) error {
	req.UpdatedById = requestUserId
	update := builders.NewUpdateDocBuilder().WithSetFields(req).Build()
	_, err := db.UpdateOne(s.Collection, oldPost.Id, update)
	return err
}

func (s *PostService) GetMany(posts *[]models.Post, organizationId string, query models.GetManyQuery, searchTerm string) (bool, int64, error) {
	filter := bson.D{{Key: "organizationId", Value: organizationId}}
	if query.OrderBy != constants.EmptyString {
		sortOptions := bson.D{{Key: query.OrderBy, Value: db.GetSortType(query.SortType)}}
		return db.FindMany(s.Collection, filter, query.Limit, query.Offset, posts, nil, sortOptions)
	}

	if searchTerm != constants.EmptyString {
		filter = append(filter, bson.E{Key: "name", Value: bson.M{"$regex": searchTerm, "$options": "i"}})
	}

	return db.FindMany(s.Collection, filter, query.Limit, query.Offset, posts, nil, nil)
}

func (s *PostService) Delete(id string) error {
	_, err := db.DeleteById(s.Collection, id)
	return err
}

func (s *PostService) GetByIds(posts *[]models.Post, ids []string, organizationId string, query models.GetManyQuery) (bool, int64, error) {
	objectIds, err := db.ConvertToObjectIds(ids)
	if err != nil {
		return false, 0, err
	}
	filter := bson.M{"organizationId": organizationId, "_id": bson.M{"$in": objectIds}}
	return db.FindMany(s.Collection, filter, query.Limit, query.Offset, posts, nil, nil)
}

func (s *PostService) DoesExist(id, organizationId string) (bool, error) {
	return db.DoesEntityExist(s.Collection, id, organizationId, new(models.Post))
}

func (s *PostService) DeleteAllForOrganization(organizationId string) error {
	filter := bson.D{{Key: "organizationId", Value: organizationId}}
	_, err := db.DeleteByMany(s.Collection, filter)
	return err
}
