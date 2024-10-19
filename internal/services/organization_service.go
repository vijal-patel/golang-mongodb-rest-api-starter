package services

import (
	"golang-mongodb-rest-api-starter/internal/constants"
	"golang-mongodb-rest-api-starter/internal/db"
	"golang-mongodb-rest-api-starter/internal/models"
	"golang-mongodb-rest-api-starter/internal/models/builders"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrganizationServiceWrapper interface {
	Create(organization *models.Organization)
	Delete(organization *models.Organization)
	Update(organization *models.Organization, updateOrganizationRequest *models.UpdateOrganizationRequest)
}

type OrganizationService struct {
	Collection *mongo.Collection
}

func NewOrganizationService(db *mongo.Client, dbName string) *OrganizationService {
	return &OrganizationService{
		Collection: db.Database(dbName).Collection(constants.OrganizationCollection),
	}
}

func (s *OrganizationService) Create(organization *models.Organization) error {
	res, err := db.InsertOne(s.Collection, &organization)
	organization.Id = res.InsertedID.(primitive.ObjectID).Hex()
	return err
}

func (s *OrganizationService) Delete(id string) error {
	_, err := db.DeleteById(s.Collection, id)
	return err
}

func (s *OrganizationService) Update(id string, updatedOrganization *models.UpdateOrganizationRequest) error {
	update := builders.NewUpdateDocBuilder().WithSetFields(updatedOrganization).Build()
	_, err := db.UpdateOne(s.Collection, id, update)
	return err
}

func (s *OrganizationService) GetMany(organizations *[]models.Organization, filter interface{}, limit int64, offset int64) (bool, int64, error) {
	return db.FindMany(s.Collection, filter, limit, offset, organizations, nil, nil)
}

func (s *OrganizationService) GetById(organization *models.Organization, id string) error {
	return db.FindById(s.Collection, id, organization, nil)
}

func (s *OrganizationService) DoesExist(id, organizationId string) (bool, error) {
	return db.DoesEntityExist(s.Collection, id, organizationId, new(models.Organization))
}
