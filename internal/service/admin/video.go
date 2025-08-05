package admin

import (
	"context"
	"errors"
	"github.com/langowen/bodybalance-backend/internal/entities/admin"
	"github.com/langowen/bodybalance-backend/pkg/lib/logger/sl"
	"github.com/theartofdevel/logging"
	"strings"
)

func (s *ServiceAdmin) AddVideo(ctx context.Context, req *admin.Video) (int64, error) {
	const op = "service.AddVideo"

	err := validVideo(req)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrVideoInvalidName):
			logging.L(ctx).Warn("empty required video name", "op", op)
			return 0, admin.ErrVideoInvalidName
		case errors.Is(err, admin.ErrVideoInvalidURL):
			logging.L(ctx).Warn("empty required video URL", "op", op)
			return 0, admin.ErrVideoInvalidURL
		case errors.Is(err, admin.ErrVideoInvalidImgURL):
			logging.L(ctx).Warn("empty required video ImgURL", "op", op)
			return 0, admin.ErrVideoInvalidImgURL
		case errors.Is(err, admin.ErrVideoInvalidCategory):
			logging.L(ctx).Warn("empty required video CategoryIDs", "op", op)
			return 0, admin.ErrVideoInvalidCategory
		case errors.Is(err, admin.ErrVideoURLPattern):
			logging.L(ctx).Warn("invalid video URL pattern", "url", req.URL, "op", op)
			return 0, admin.ErrVideoURLPattern
		case errors.Is(err, admin.ErrVideoSuspiciousPattern):
			logging.L(ctx).Warn("suspicious pattern in video URL", "url", req.URL, "op", op)
			return 0, admin.ErrVideoSuspiciousPattern
		case errors.Is(err, admin.ErrVideoImgPattern):
			logging.L(ctx).Warn("invalid video ImgURL pattern", "imgurl", req.ImgURL, "op", op)
			return 0, admin.ErrVideoImgPattern
		case errors.Is(err, admin.ErrVideoImgSuspiciousPattern):
			logging.L(ctx).Warn("suspicious pattern in video ImgURL", "imgurl", req.ImgURL, "op", op)
			return 0, admin.ErrVideoImgSuspiciousPattern
		}
	}

	video, err := s.db.AddVideo(ctx, req)
	if err != nil {
		logging.L(ctx).Error("failed to add video", "op", op, "video", video, sl.Err(err))
		return 0, admin.ErrVideoSaveFailed
	}

	if s.cfg.Redis.Enable == true {
		go s.removeCache(ctx, op)
	}

	return video, nil
}

func (s *ServiceAdmin) GetVideo(ctx context.Context, id int64) (*admin.Video, error) {
	const op = "service.GetVideo"

	if id <= 0 {
		logging.L(ctx).Error("invalid video ID", "id", id, "op", op)
		return nil, admin.ErrVideoInvalidID
	}

	video, err := s.db.GetVideo(ctx, id)
	if err != nil {
		if errors.Is(err, admin.ErrVideoNotFound) {
			logging.L(ctx).Warn("video not found", "id", id, "op", op)
			return nil, admin.ErrVideoNotFound
		}
		logging.L(ctx).Error("failed to get video", "op", op, "video_id", id, sl.Err(err))
		return nil, admin.ErrFailedGetVideo
	}

	return video, nil
}

func (s *ServiceAdmin) GetVideos(ctx context.Context) ([]admin.Video, error) {
	const op = "service.GetVideos"

	videos, err := s.db.GetVideos(ctx)
	if err != nil {
		if errors.Is(err, admin.ErrVideoNotFound) {
			logging.L(ctx).Warn("no videos found", "op", op)
			return nil, admin.ErrVideoNotFound
		}
		logging.L(ctx).Error("failed to get videos", "op", op, sl.Err(err))
		return nil, admin.ErrFailedGetVideo
	}

	return videos, nil
}

func (s *ServiceAdmin) UpdateVideo(ctx context.Context, req *admin.Video) error {
	const op = "service.UpdateVideo"

	if req.ID <= 0 {
		logging.L(ctx).Error("invalid video ID", "id", req.ID, "op", op)
		return admin.ErrVideoInvalidID
	}

	err := validVideo(req)
	if err != nil {
		switch {
		case errors.Is(err, admin.ErrVideoInvalidName):
			logging.L(ctx).Warn("empty required video name", "op", op)
			return admin.ErrVideoInvalidName
		case errors.Is(err, admin.ErrVideoInvalidURL):
			logging.L(ctx).Warn("empty required video URL", "op", op)
			return admin.ErrVideoInvalidURL
		case errors.Is(err, admin.ErrVideoInvalidImgURL):
			logging.L(ctx).Warn("empty required video ImgURL", "op", op)
			return admin.ErrVideoInvalidImgURL
		case errors.Is(err, admin.ErrVideoInvalidCategory):
			logging.L(ctx).Warn("empty required video CategoryIDs", "op", op)
			return admin.ErrVideoInvalidCategory
		case errors.Is(err, admin.ErrVideoURLPattern):
			logging.L(ctx).Warn("invalid video URL pattern", "url", req.URL, "op", op)
			return admin.ErrVideoURLPattern
		case errors.Is(err, admin.ErrVideoSuspiciousPattern):
			logging.L(ctx).Warn("suspicious pattern in video URL", "url", req.URL, "op", op)
			return admin.ErrVideoSuspiciousPattern
		case errors.Is(err, admin.ErrVideoImgPattern):
			logging.L(ctx).Warn("invalid video ImgURL pattern", "imgurl", req.ImgURL, "op", op)
			return admin.ErrVideoImgPattern
		case errors.Is(err, admin.ErrVideoImgSuspiciousPattern):
			logging.L(ctx).Warn("suspicious pattern in video ImgURL", "imgurl", req.ImgURL, "op", op)
			return admin.ErrVideoImgSuspiciousPattern
		}
	}

	err = s.db.UpdateVideo(ctx, req)
	if err != nil {
		if errors.Is(err, admin.ErrVideoNotFound) {
			logging.L(ctx).Warn("video not found", "op", op, "video_id", req.ID, sl.Err(err))
			return admin.ErrVideoNotFound
		}
		logging.L(ctx).Error("failed to update video", "op", op, "video_id", req.ID, sl.Err(err))
		return admin.ErrVideoUpdateFailed
	}

	if s.cfg.Redis.Enable == true {
		go s.removeCache(ctx, op)
	}

	return nil
}

func (s *ServiceAdmin) DeleteVideo(ctx context.Context, id int64) error {
	const op = "service.DeleteVideo"

	if id <= 0 {
		logging.L(ctx).Error("invalid video ID", "id", id, "op", op)
		return admin.ErrVideoInvalidID
	}

	err := s.db.DeleteVideo(ctx, id)
	if err != nil {
		if errors.Is(err, admin.ErrVideoNotFound) {
			logging.L(ctx).Warn("video not found", "op", op, "video_id", id, sl.Err(err))
			return admin.ErrVideoNotFound
		}
		logging.L(ctx).Error("failed to delete video", "op", op, "video_id", id, sl.Err(err))
		return admin.ErrVideoDeleteFailed
	}

	if s.cfg.Redis.Enable == true {
		go s.removeCache(ctx, op)
	}

	return nil
}

// validUser проверят входящие данные на валидность
func validVideo(req *admin.Video) error {
	switch {
	case req.Name == "":
		return admin.ErrVideoInvalidName
	case req.URL == "":
		return admin.ErrVideoInvalidURL
	case req.ImgURL == "":
		return admin.ErrVideoInvalidImgURL
	case len(req.Categories) == 0:
		return admin.ErrVideoInvalidCategory
	case !validFilePattern.MatchString(req.URL):
		return admin.ErrVideoURLPattern
	case !validFilePattern.MatchString(req.ImgURL):
		return admin.ErrVideoImgPattern
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(req.URL, pattern) {
			return admin.ErrVideoSuspiciousPattern
		}
	}

	for _, pattern := range suspiciousPatterns {
		if strings.Contains(req.ImgURL, pattern) {
			return admin.ErrVideoImgSuspiciousPattern
		}
	}

	return nil
}
