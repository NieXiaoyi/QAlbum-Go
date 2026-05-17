package service

import (
	"fmt"
	"strconv"

	"qalbum-server/pkg/dao"
	"qalbum-server/pkg/model"
	"qalbum-server/pkg/phash"
)

type DuplicateService struct {
	fileDAO *dao.FileDAO
	spaceDAO *dao.SpaceDAO
}

func NewDuplicateService(fileDAO *dao.FileDAO, spaceDAO *dao.SpaceDAO) *DuplicateService {
	return &DuplicateService{
		fileDAO: fileDAO,
		spaceDAO: spaceDAO,
	}
}

type DuplicateGroup struct {
	Files      []*model.File
	Similarity int
}

func (s *DuplicateService) Detect(spaceID int, userID int, threshold int) ([]*DuplicateGroup, error) {
	isMember, err := s.spaceDAO.IsMember(spaceID, userID)
	if err != nil || !isMember {
		return nil, fmt.Errorf("not a member")
	}

	images, err := s.fileDAO.GetImagesBySpace(spaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get images: %w", err)
	}

	var filesWithHash []*phash.FileWithPHash
	for _, img := range images {
		if img.PHash != nil {
			hash, err := strconv.ParseUint(*img.PHash, 16, 64)
			if err != nil {
				continue
			}
			filesWithHash = append(filesWithHash, &phash.FileWithPHash{
				ID:    img.ID,
				PHash: hash,
				File:  img,
			})
		}
	}

	groups := s.clusterByHammingDistance(filesWithHash, threshold)
	
	var result []*DuplicateGroup
	for _, group := range groups {
		if len(group.Files) > 1 {
			var files []*model.File
			for _, f := range group.Files {
				files = append(files, f.File.(*model.File))
			}
			result = append(result, &DuplicateGroup{
				Files:      files,
				Similarity: group.Similarity,
			})
		}
	}

	return result, nil
}

func (s *DuplicateService) clusterByHammingDistance(files []*phash.FileWithPHash, threshold int) []*phash.Group {
	var groups []*phash.Group
	visited := make(map[int]bool)

	for i, file1 := range files {
		if visited[file1.ID] {
			continue
		}

		var similar []*phash.FileWithPHash
		similar = append(similar, file1)
		visited[file1.ID] = true
		maxDistance := 0

		for j := i + 1; j < len(files); j++ {
			file2 := files[j]
			if visited[file2.ID] {
				continue
			}

			distance := phash.HammingDistance(file1.PHash, file2.PHash)
			if distance <= threshold {
				similar = append(similar, file2)
				visited[file2.ID] = true
				if distance > maxDistance {
					maxDistance = distance
				}
			}
		}

		if len(similar) > 1 {
			groups = append(groups, &phash.Group{
				Files:      similar,
				Similarity: maxDistance,
			})
		}
	}

	return groups
}
