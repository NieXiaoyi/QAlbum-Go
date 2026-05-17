package service

import (
	"fmt"
	"qalbum-server/pkg/dao"
	"qalbum-server/pkg/model"
)

type AlbumService struct {
	albumDAO   *dao.AlbumDAO
	spaceDAO   *dao.SpaceDAO
}

func NewAlbumService(albumDAO *dao.AlbumDAO, spaceDAO *dao.SpaceDAO) *AlbumService {
	return &AlbumService{
		albumDAO:   albumDAO,
		spaceDAO:   spaceDAO,
	}
}

func (s *AlbumService) Create(spaceID int, name string, userID int) (*model.Album, error) {
	isMember, err := s.spaceDAO.IsMember(spaceID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("not a member")
	}

	return s.albumDAO.Create(spaceID, name)
}

func (s *AlbumService) GetByID(id int, userID int) (*model.Album, error) {
	album, err := s.albumDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	isMember, err := s.spaceDAO.IsMember(album.SpaceID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("not a member")
	}

	return album, nil
}

func (s *AlbumService) ListBySpace(spaceID int, userID int) ([]*model.Album, error) {
	isMember, err := s.spaceDAO.IsMember(spaceID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("not a member")
	}

	return s.albumDAO.ListBySpace(spaceID)
}

func (s *AlbumService) Update(id int, name string, userID int) (*model.Album, error) {
	album, err := s.albumDAO.GetByID(id)
	if err != nil {
		return nil, err
	}

	isMember, err := s.spaceDAO.IsMember(album.SpaceID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("not a member")
	}

	return s.albumDAO.Update(id, name)
}

func (s *AlbumService) SoftDelete(id int, userID int) error {
	album, err := s.albumDAO.GetByID(id)
	if err != nil {
		return err
	}

	isMember, err := s.spaceDAO.IsMember(album.SpaceID, userID)
	if err != nil || !isMember {
		return fmt.Errorf("not a member")
	}

	return s.albumDAO.SoftDelete(id)
}
