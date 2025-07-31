package api

import (
	"context"
	"errors"
	"github.com/langowen/bodybalance-backend/deploy/config"
	"github.com/langowen/bodybalance-backend/internal/adapter/storage"
	"github.com/langowen/bodybalance-backend/internal/entities/api"
	"github.com/langowen/bodybalance-backend/internal/lib/logger/sl"
	mwMetrics "github.com/langowen/bodybalance-backend/internal/port/http-server/middleware/metrics"
	"github.com/redis/go-redis/v9"
	"github.com/theartofdevel/logging"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ServiceApi struct {
	cfg *config.Config
	db  SqlStorageApi
	rdb CacheStorageApi
}

func NewServiceApi(cfg *config.Config, db SqlStorageApi, rdb CacheStorageApi) *ServiceApi {
	return &ServiceApi{
		cfg: cfg,
		db:  db,
		rdb: rdb,
	}
}

func (s *ServiceApi) GetTypeByAccount(ctx context.Context, username string) (*api.Account, string, error) {
	const op = "service.GetTypeByAccount"

	if username == "" {
		logging.L(ctx).Error("Username is empty", "op", op)
		return nil, "", api.ErrEmptyUsername
	}

	account := api.Account{
		Username: username,
	}

	if s.cfg.Redis.Enable == true {
		res, err := s.rdb.GetAccount(ctx, &account)
		if err == nil {
			logging.L(ctx).Debug("serving from cache", "account_type", res.ContentType.Name)
			return res, mwMetrics.SourceRedis, nil
		}

		if errors.Is(err, redis.Nil) {
			logging.L(ctx).Debug("account not found in redis cache", sl.Err(err))
		} else {
			logging.L(ctx).Error("failed to get account from redis", sl.Err(err))
		}
	}

	res, err := s.db.CheckAccount(ctx, &account)
	if err != nil {
		if errors.Is(err, storage.ErrAccountNotFound) {
			return nil, "", err
		}

		logging.L(ctx).Error("storage get error", sl.Err(err))
		return nil, "", api.ErrStorageServerError
	}

	if s.cfg.Redis.Enable {
		go func() {
			ctxRedis, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err = s.rdb.SetAccount(ctxRedis, res); err != nil {
				logging.L(ctx).Warn("failed to cache account in redis", sl.Err(err))
			}
		}()
	}

	return res, mwMetrics.SourceSQL, nil
}

func (s *ServiceApi) GetCategoriesByType(ctx context.Context, contentType string) ([]api.Category, string, error) {
	const op = "service.getCategoriesByType"

	if contentType == "" {
		logging.L(ctx).Error("Content type ID is empty", "op", op)
		return nil, "", api.ErrEmptyTypeID
	}

	typeID, err := strconv.ParseInt(contentType, 10, 64)
	if err != nil {
		logging.L(ctx).Error("invalid type ID", "op", op, sl.Err(err))
		return nil, "", api.ErrTypeInvalid
	}

	if s.cfg.Redis.Enable {
		categories, err := s.rdb.GetCategories(ctx, typeID)
		if err == nil && categories != nil {
			logging.L(ctx).Debug("categories fetched from redis cache", "op", op)
			return categories, mwMetrics.SourceRedis, nil
		}

		if err != nil {
			if errors.Is(err, redis.Nil) {
				logging.L(ctx).Debug("categories not found in redis cache", sl.Err(err))
			} else {
				logging.L(ctx).Error("failed to get categories from redis", sl.Err(err))
			}
		}
	}

	categories, err := s.db.GetCategories(ctx, typeID)
	if err != nil {
		if errors.Is(err, storage.ErrContentTypeNotFound) {
			logging.L(ctx).Debug("content type not found", sl.Err(err))
			return nil, "", err
		}

		logging.L(ctx).Error("failed to get categories from DB", sl.Err(err))
		return nil, "", api.ErrStorageServerError
	}

	if s.cfg.Redis.Enable && categories != nil {
		go func() {
			ctxRedis, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err = s.rdb.SetCategories(ctxRedis, typeID, categories); err != nil {
				logging.L(ctx).Error("failed to cache categories in redis", sl.Err(err))
			}
		}()
	}

	return categories, mwMetrics.SourceSQL, nil
}

func (s *ServiceApi) GetVideo(ctx context.Context, videoStr string) (*api.Video, string, error) {
	const op = "service.GetVideo"

	if videoStr == "" {
		logging.L(ctx).Error("Video id is empty", "op", op)
		return nil, "", api.ErrEmptyVideoID
	}

	videoID, err := strconv.ParseInt(videoStr, 10, 64)
	if err != nil {
		logging.L(ctx).Error("Invalid video ID", "op", op, sl.Err(err))
		return nil, "", api.ErrInvalidVideoID
	}

	if s.cfg.Redis.Enable {
		video, err := s.rdb.GetVideo(ctx, videoID)
		if err == nil && video != nil {
			logging.L(ctx).Debug("video fetched from redis cache")
			return video, mwMetrics.SourceRedis, nil
		}

		if err != nil {
			if errors.Is(err, redis.Nil) {
				logging.L(ctx).Debug("video not found in redis cache", sl.Err(err))
			} else {
				logging.L(ctx).Error("failed to get video from redis", sl.Err(err))
			}
		}
	}

	video, err := s.db.GetVideo(ctx, videoID)
	if err != nil {
		if errors.Is(err, storage.ErrVideoNotFound) {
			logging.L(ctx).Debug("video not found in DB", sl.Err(err))
			return nil, "", storage.ErrVideoNotFound
		}

		logging.L(ctx).Error("failed to get video", sl.Err(err))
		return nil, "", api.ErrStorageServerError
	}

	if s.cfg.Redis.Enable && video != nil {
		go func() {
			ctxRedis, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err = s.rdb.SetVideo(ctxRedis, videoID, video); err != nil {
				logging.L(ctx).Warn("failed to cache video in redis", sl.Err(err))
			}
		}()
	}

	return video, mwMetrics.SourceSQL, nil
}

func (s *ServiceApi) GetVideosByCategoryAndType(ctx context.Context, contentType, category string) ([]api.Video, string, error) {
	const op = "service.GetVideosByCategoryAndType"

	if category == "" {
		logging.L(ctx).Error("Category is empty", "op", op)
		return nil, "", api.ErrEmptyCategoryID
	}

	if contentType == "" {
		logging.L(ctx).Error("Content type is empty", "op", op)
		return nil, "", api.ErrEmptyTypeID
	}

	typeID, err := strconv.ParseInt(contentType, 10, 64)
	if err != nil {
		logging.L(ctx).Error("Invalid type ID", "op", op, sl.Err(err))
		return nil, "", api.ErrTypeInvalid
	}

	catID, err := strconv.ParseInt(category, 10, 64)
	if err != nil {
		logging.L(ctx).Error("Invalid category ID", "op", op, sl.Err(err))
		return nil, "", api.ErrCategoryInvalid
	}

	if s.cfg.Redis.Enable == true {
		videos, err := s.rdb.GetVideosByCategoryAndType(ctx, typeID, catID)
		if err == nil && videos != nil {
			logging.L(ctx).Debug("videos fetched from redis cache")
			return videos, mwMetrics.SourceRedis, nil
		}

		if err != nil {
			if errors.Is(err, redis.Nil) {
				logging.L(ctx).Debug("videos not found in redis cache", sl.Err(err))
			} else {
				logging.L(ctx).Error("failed to get videos from redis", sl.Err(err))
			}
		}
	}

	videos, err := s.db.GetVideosByCategoryAndType(ctx, typeID, catID)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrContentTypeNotFound):
			logging.L(ctx).Warn("content type not found", sl.Err(err))
			return nil, "", err
		case errors.Is(err, storage.ErrNoCategoriesFound):
			logging.L(ctx).Warn("no categories found", sl.Err(err))
			return nil, "", err
		case errors.Is(err, storage.ErrVideoNotFound):
			logging.L(ctx).Warn("video not found", sl.Err(err))
			return nil, "", err
		default:
			logging.L(ctx).Error("Failed to get videos", sl.Err(err))
			return nil, "", api.ErrStorageServerError
		}
	}

	if s.cfg.Redis.Enable == true {
		go func() {
			ctxRedis, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			if err = s.rdb.SetVideosByCategoryAndType(ctxRedis, typeID, catID, videos); err != nil {
				logging.L(ctx).Warn("failed to set videos cache", sl.Err(err))
			}
		}()
	}

	return videos, mwMetrics.SourceSQL, nil
}

func (s *ServiceApi) Feedback(ctx context.Context, feedback *api.Feedback) error {
	const op = "service.Feedback"

	if feedback.Message == "" {
		logging.L(ctx).Warn("message is required", "op", op)
		return api.ErrEmptyMessage
	}

	if feedback.Email != "" {
		if !isValidEmail(feedback.Email) {
			logging.L(ctx).Warn("invalid email format", "email", feedback.Email, "op", op)
			return api.ErrInvalidEmail
		}
	}

	if feedback.Telegram != "" {
		if !isValidTelegram(feedback.Telegram) {
			logging.L(ctx).Warn("invalid telegram format", "telegram", feedback.Telegram, "op", op)
			return api.ErrInvalidTelegram
		}
	}

	if feedback.Email == "" && feedback.Telegram == "" {
		logging.L(ctx).Warn("no contact method provided", "op", op)
		return api.ErrEmptyTelegramOrEmail
	}

	err := s.db.Feedback(ctx, feedback)
	if err != nil {
		logging.L(ctx).Error("feedback save error", sl.Err(err))
		return api.ErrStorageServerError
	}

	return nil
}

func isValidEmail(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isValidTelegram(telegram string) bool {
	if !strings.HasPrefix(telegram, "@") {
		return false
	}
	if len(telegram) < 6 || len(telegram) > 33 { // @ + 5-32 символа
		return false
	}
	telegramRegex := regexp.MustCompile(`^@[a-zA-Z0-9_]+$`)
	return telegramRegex.MatchString(telegram)
}
