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

	"golang.org/x/crypto/bcrypt"
)

type UserServiceWrapper interface {
	Register(request *models.RegisterRequest) error
}

type UserService struct {
	Collection *mongo.Collection
}

func NewUserService(db *mongo.Client, dbName string) *UserService {
	return &UserService{Collection: db.Database(dbName).Collection(constants.UserCollection)}
}

func (s *UserService) Register(request *models.RegisterRequest, confirmOtp string) (*models.User, error) {
	user := models.User{}
	encryptedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(request.Password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return nil, err
	}
	user.ConfirmOtp = confirmOtp
	user.Password = string(encryptedPassword)
	user.Name = request.UserName
	user.Email = request.Email
	user.Confirmed = false
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Roles = []string{constants.BaseUserRole, constants.OrganizationAdminRole}
	user.Permissions = []string{}
	res, err := db.InsertOne(s.Collection, &user)
	user.Id = res.InsertedID.(primitive.ObjectID).Hex()
	return &user, err
}

func (s *UserService) CreateFromInvite(req *models.CreateUserRequest, requestUserId, organizationId, loginOtp string) (*models.User, error) {
	user := models.User{}
	user.Name = req.Name
	user.Email = req.Email
	user.Confirmed = true
	user.Password = loginOtp
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.CreatedById = requestUserId
	user.Roles = []string{req.Role}
	user.OrganizationId = organizationId
	res, err := db.InsertOne(s.Collection, &user)
	user.Permissions = []string{}
	user.Id = res.InsertedID.(primitive.ObjectID).Hex()
	return &user, err
}

func (s *UserService) BulkCreateFromInvite(req *models.BulkCreateUsersRequest, requestUserId, organizationId string) error {
	var users []interface{}
	for _, reqUser := range req.Users {
		user := models.User{}
		user.Name = reqUser.Name
		user.Email = reqUser.Email
		user.Confirmed = true
		user.Password = reqUser.HashedLoginOtp
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		user.CreatedById = requestUserId
		user.Roles = []string{reqUser.Role}
		user.OrganizationId = organizationId
		user.Permissions = []string{}

		users = append(users, user)
	}
	_, err := db.InsertMany(s.Collection, users)
	return err
}

func (s *UserService) DoesExist(id, organizationId string) (bool, error) {
	return db.DoesEntityExist(s.Collection, id, organizationId, new(models.User))
}

func (s *UserService) ValidateConfirmOtp(id, otp string) (bool, error) {
	user := new(models.User)
	if err := db.FindById(s.Collection, id, user, bson.D{{Key: "confirmOtp", Value: 1}}); err != nil {
		return false, err
	}
	if otp != user.ConfirmOtp {
		return false, nil
	}
	updateOtpReq := models.UpdateUserOTPRequest{
		Confirmed:  true,
		ConfirmOtp: "-",
	}
	update := builders.NewUpdateDocBuilder().WithSetFields(updateOtpReq).Build()
	_, err := db.UpdateOne(s.Collection, id, update)
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *UserService) UpdateConfirmOtp(id, otp string) error {
	updateOtpReq := models.UpdateUserOTPRequest{
		ConfirmOtp: otp,
	}
	update := builders.NewUpdateDocBuilder().WithSetFields(updateOtpReq).Build()
	if _, err := db.UpdateOne(s.Collection, id, update); err != nil {
		return err
	}

	return nil
}

func (s *UserService) UpdatePassword(id, newPassword string) error {
	encryptedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(newPassword),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return err
	}
	req := models.UpdatePasswordRequest{
		Password: string(encryptedPassword),
	}
	update := builders.NewUpdateDocBuilder().WithSetFields(req).Build()
	if _, err := db.UpdateOne(s.Collection, id, update); err != nil {
		return err
	}

	return nil
}

func (s *UserService) GetByEmail(user *models.User, email string, organizationId string) error {
	filter := bson.D{{Key: "email", Value: email}}
	if organizationId != constants.EmptyString {
		filter = append(filter, bson.E{Key: "organizationId", Value: organizationId})
	}
	return db.FindOne(s.Collection, filter, user, nil)
}

func (s *UserService) GetByEmailAnyOrg(user *models.User, email string) error {
	filter := bson.D{{Key: "email", Value: email}}
	return db.FindOne(s.Collection, filter, user, nil)
}

func (s *UserService) GetByEmailsAnyOrg(users *[]models.User, emails []string, limit, offset int64) (bool, int64, error) {
	filter := bson.D{
		{Key: "email", Value: bson.M{"$in": emails}},
	}
	return db.FindMany(s.Collection, filter, limit, offset, users, nil, nil)
}

func (s *UserService) GetById(user *models.User, id, organizationId string) error {
	objectId, _ := primitive.ObjectIDFromHex(id)
	filter := bson.D{{Key: "organizationId", Value: organizationId}, {Key: "_id", Value: objectId}}
	return db.FindOne(s.Collection, filter, user, nil)
}

func (s *UserService) GetByIdAnyOrg(user *models.User, id string) error {
	return db.FindById(s.Collection, id, user, nil)
}

func (s *UserService) Update(req *models.UpdateUserRequest, id, requestUserId, confirmOtp string) error {
	if req.Email != constants.EmptyString {
		req.Confirmed = false
		req.ConfirmOtp = confirmOtp
	}
	req.UpdatedById = requestUserId
	update := builders.NewUpdateDocBuilder().WithSetFields(req).Build()
	_, err := db.UpdateOne(s.Collection, id, update)
	return err
}

func (s *UserService) GetMany(users *[]models.User, organizationId string, query models.GetManyQuery, searchTerm string) (bool, int64, error) {
	filter := bson.D{{Key: "organizationId", Value: organizationId}}

	if searchTerm != constants.EmptyString {
		filter = append(filter, bson.E{Key: "name", Value: bson.M{"$regex": searchTerm, "$options": "i"}})
	}

	if query.OrderBy != constants.EmptyString {
		sortOptions := bson.D{{Key: query.OrderBy, Value: db.GetSortType(query.SortType)}}
		return db.FindMany(s.Collection, filter, query.Limit, query.Offset, users, nil, sortOptions)
	}
	return db.FindMany(s.Collection, filter, query.Limit, query.Offset, users, nil, nil)
}

func (s *UserService) Delete(id string) error {
	_, err := db.DeleteById(s.Collection, id)
	return err
}

func (s *UserService) GetTotalOrgAdminsInOrg(organizationId, userId string) (int64, error) {
	filter := bson.M{"organizationId": organizationId, "roles": bson.M{"$in": []string{constants.OrganizationAdminRole}}}
	return db.GetFilterCount(s.Collection, filter)
}

func (s *UserService) GetTotalUsersInOrg(organizationId string) (int64, error) {
	filter := bson.M{"organizationId": organizationId}
	return db.GetFilterCount(s.Collection, filter)
}

func (s *UserService) RevokeLoginOTP(userId string) (*mongo.UpdateResult, error) {
	update := bson.D{{Key: "confirmed", Value: true}, {Key: "loginOtp", Value: nil}}
	return db.UpdateOne(s.Collection, userId, update)
}

func (s *UserService) DeleteAllForOrganization(organizationId string) error {
	filter := bson.D{{Key: "organizationId", Value: organizationId}}
	_, err := db.DeleteByMany(s.Collection, filter)
	return err
}
