// internal/service/team_service.go
package service

import (
	"errors"

	"github.com/Axontik/comin-organization-service/internal/domain"
	"github.com/Axontik/comin-organization-service/internal/repository"
	"github.com/google/uuid"
)

type TeamService interface {
	Create(orgID uuid.UUID, req *domain.TeamRequest) (*domain.Team, error)
	GetByID(orgID, teamID uuid.UUID) (*domain.Team, error)
	Update(orgID, teamID uuid.UUID, req *domain.TeamRequest) (*domain.Team, error)
	Delete(orgID, teamID uuid.UUID) error
	List(orgID uuid.UUID) ([]domain.Team, error)
	ListByDepartment(orgID, deptID uuid.UUID) ([]domain.Team, error)
	AddMember(orgID, teamID, userID uuid.UUID, role string) error
	RemoveMember(orgID, teamID, userID uuid.UUID) error
	ListMembers(orgID, teamID uuid.UUID) ([]domain.TeamMember, error)
	UpdateTeamLead(orgID, teamID, leadID uuid.UUID) error
}

type teamService struct {
	teamRepo repository.TeamRepository
}

func NewTeamService(teamRepo repository.TeamRepository) TeamService {
	return &teamService{
		teamRepo: teamRepo,
	}
}

func (s *teamService) Create(orgID uuid.UUID, req *domain.TeamRequest) (*domain.Team, error) {
	team := &domain.Team{
		OrganizationID: orgID,
		DepartmentID:   req.DepartmentID,
		Name:           req.Name,
		Description:    req.Description,
		LeadID:         req.LeadID,
		Status:         "active",
	}

	if err := s.teamRepo.Create(team); err != nil {
		return nil, err
	}

	// If team lead is specified, add them as a member with "lead" role
	if team.LeadID != nil {
		member := &domain.TeamMember{
			TeamID: team.ID,
			UserID: *team.LeadID,
			Role:   "lead",
		}
		if err := s.teamRepo.AddMember(member); err != nil {
			return nil, err
		}
	}

	return team, nil
}

func (s *teamService) GetByID(orgID, teamID uuid.UUID) (*domain.Team, error) {
	team, err := s.teamRepo.GetByID(teamID)
	if err != nil {
		return nil, err
	}

	if team.OrganizationID != orgID {
		return nil, errors.New("team not found in organization")
	}

	return team, nil
}

func (s *teamService) Update(orgID, teamID uuid.UUID, req *domain.TeamRequest) (*domain.Team, error) {
	team, err := s.GetByID(orgID, teamID)
	if err != nil {
		return nil, err
	}

	team.Name = req.Name
	team.Description = req.Description
	team.DepartmentID = req.DepartmentID

	// Handle team lead changes
	if req.LeadID != nil && (team.LeadID == nil || *team.LeadID != *req.LeadID) {
		err := s.UpdateTeamLead(orgID, teamID, *req.LeadID)
		if err != nil {
			return nil, err
		}
		team.LeadID = req.LeadID
	}

	if err := s.teamRepo.Update(team); err != nil {
		return nil, err
	}

	return team, nil
}

func (s *teamService) Delete(orgID, teamID uuid.UUID) error {
	_, err := s.GetByID(orgID, teamID)
	if err != nil {
		return err
	}

	return s.teamRepo.Delete(teamID)
}

func (s *teamService) List(orgID uuid.UUID) ([]domain.Team, error) {
	return s.teamRepo.List(orgID)
}

func (s *teamService) ListByDepartment(orgID, deptID uuid.UUID) ([]domain.Team, error) {
	return s.teamRepo.ListByDepartment(deptID)
}

func (s *teamService) AddMember(orgID, teamID, userID uuid.UUID, role string) error {
	// Verify team belongs to organization
	if _, err := s.GetByID(orgID, teamID); err != nil {
		return err
	}

	member := &domain.TeamMember{
		TeamID: teamID,
		UserID: userID,
		Role:   role,
	}

	return s.teamRepo.AddMember(member)
}

// func (s *teamService) Create(orgID uuid.UUID, req *domain.TeamRequest) (*domain.Team, error) {
// 	team := &domain.Team{
// 		OrganizationID: orgID,
// 		DepartmentID:   req.DepartmentID,
// 		Name:           req.Name,
// 		Description:    req.Description,
// 		LeadID:         req.LeadID,
// 		Status:         "active",
// 	}

// 	if err := s.teamRepo.Create(team); err != nil {
// 		return nil, err
// 	}

// 	// If team lead is specified, add them as a member with "lead" role
// 	if team.LeadID != nil {
// 		member := &domain.TeamMember{
// 			TeamID: team.ID,
// 			UserID: *team.LeadID,
// 			Role:   "lead",
// 		}
// 		if err := s.teamRepo.AddMember(member); err != nil {
// 			return nil, err
// 		}
// 	}

// 	return team, nil
// }

func (s *teamService) RemoveMember(orgID, teamID, userID uuid.UUID) error {
	// Verify team belongs to organization
	if _, err := s.GetByID(orgID, teamID); err != nil {
		return err
	}

	return s.teamRepo.RemoveMember(teamID, userID)
}

func (s *teamService) ListMembers(orgID, teamID uuid.UUID) ([]domain.TeamMember, error) {
	// Verify team belongs to organization
	if _, err := s.GetByID(orgID, teamID); err != nil {
		return nil, err
	}

	return s.teamRepo.ListMembers(teamID)
}

func (s *teamService) UpdateTeamLead(orgID, teamID, leadID uuid.UUID) error {
	team, err := s.GetByID(orgID, teamID)
	if err != nil {
		return err
	}

	// Remove old team lead role if exists
	if team.LeadID != nil {
		oldLead := &domain.TeamMember{
			TeamID: teamID,
			UserID: *team.LeadID,
			Role:   "member",
		}
		if err := s.teamRepo.UpdateMember(oldLead); err != nil {
			return err
		}
	}

	// Add or update new team lead
	newLead := &domain.TeamMember{
		TeamID: teamID,
		UserID: leadID,
		Role:   "lead",
	}

	return s.teamRepo.AddMember(newLead)
}
