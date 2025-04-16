// internal/handler/team_handler.go
package handler

import (
	"net/http"

	"github.com/Axontik/comin-organization-service/internal/domain"
	"github.com/Axontik/comin-organization-service/internal/errors"
	"github.com/Axontik/comin-organization-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type TeamHandler struct {
	teamService service.TeamService
}

func NewTeamHandler(teamService service.TeamService) *TeamHandler {
	return &TeamHandler{teamService: teamService}
}

// @Summary Create team
// @Description Create a new team in an organization
// @Tags teams
// @Accept json
// @Produce json
// @Param org_id path string true "Organization ID"
// @Param team body domain.TeamRequest true "Team details"
// @Success 201 {object} domain.Team
// @Router /organizations/{org_id}/teams [post]
func (h *TeamHandler) CreateTeam(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("org_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid organization id"))
		return
	}

	var req domain.TeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	team, err := h.teamService.Create(orgID, &req)
	if err != nil {
		c.Error(errors.NewInternalServerError("failed to create team"))
		return
	}

	c.JSON(http.StatusCreated, team)
}

// @Summary Get team by ID
// @Tags teams
// @Produce json
// @Param org_id path string true "Organization ID"
// @Param id path string true "Team ID"
// @Success 200 {object} domain.Team
// @Router /organizations/{org_id}/teams/{team_id} [get]
func (h *TeamHandler) GetTeamByID(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("org_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid organization id"))
		return
	}

	teamID, err := uuid.Parse(c.Param("team_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid team id"))
		return
	}

	team, err := h.teamService.GetByID(orgID, teamID)
	if err != nil {
		c.Error(errors.NewNotFoundError("team not found"))
		return
	}

	c.JSON(http.StatusOK, team)
}

func (h *TeamHandler) List(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("org_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid organization id"))
		return
	}

	teams, err := h.teamService.List(orgID)
	if err != nil {
		c.Error(errors.NewInternalServerError("failed to list teams"))
		return
	}

	c.JSON(http.StatusOK, teams)
}

func (h *TeamHandler) UpdateTeam(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("org_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid organization id"))
		return
	}

	teamID, err := uuid.Parse(c.Param("team_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid team id"))
		return
	}

	var req struct {
		Name         string     `json:"name"`
		Description  string     `json:"description"`
		DepartmentID *uuid.UUID `json:"department_id"`
		LeadID       *uuid.UUID `json:"lead_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	team, err := h.teamService.Update(orgID, teamID, &domain.TeamRequest{
		Name:         req.Name,
		Description:  req.Description,
		DepartmentID: req.DepartmentID,
		LeadID:       req.LeadID,
	})
	if err != nil {
		c.Error(errors.NewInternalServerError("failed to update team"))
		return
	}

	c.JSON(http.StatusOK, team)
}

func (h *TeamHandler) DeleteTeam(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("org_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid organization id"))
		return
	}

	teamID, err := uuid.Parse(c.Param("team_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid team id"))
		return
	}

	if err := h.teamService.Delete(orgID, teamID); err != nil {
		c.Error(errors.NewInternalServerError("failed to delete team"))
		return
	}

	c.Status(http.StatusNoContent)
}

// Member management handlers
func (h *TeamHandler) AddMember(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("org_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid organization id"))
		return
	}

	teamID, err := uuid.Parse(c.Param("team_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid team id"))
		return
	}

	var req struct {
		UserID uuid.UUID `json:"user_id" binding:"required"`
		Role   string    `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	err = h.teamService.AddMember(orgID, teamID, req.UserID, req.Role)
	if err != nil {
		c.Error(errors.NewInternalServerError("failed to add team member"))
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Team member added successfully",
	})
}

func (h *TeamHandler) DeleteMember(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("org_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid organization id"))
		return
	}

	teamID, err := uuid.Parse(c.Param("team_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid team id"))
		return
	}

	userID, err := uuid.Parse(c.Param("user_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid user id"))
		return
	}

	if err := h.teamService.RemoveMember(orgID, teamID, userID); err != nil {
		c.Error(errors.NewInternalServerError("failed to remove team member"))
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *TeamHandler) ListMembers(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("org_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid organization id"))
		return
	}

	teamID, err := uuid.Parse(c.Param("team_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid team id"))
		return
	}

	members, err := h.teamService.ListMembers(orgID, teamID)
	if err != nil {
		c.Error(errors.NewInternalServerError("failed to list team members"))
		return
	}

	c.JSON(http.StatusOK, members)
}

func (h *TeamHandler) UpdateTeamLead(c *gin.Context) {
	orgID, err := uuid.Parse(c.Param("org_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid organization id"))
		return
	}

	teamID, err := uuid.Parse(c.Param("team_id"))
	if err != nil {
		c.Error(errors.NewBadRequestError("invalid team id"))
		return
	}

	var req struct {
		LeadID uuid.UUID `json:"lead_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(errors.NewBadRequestError(err.Error()))
		return
	}

	if err := h.teamService.UpdateTeamLead(orgID, teamID, req.LeadID); err != nil {
		c.Error(errors.NewInternalServerError("failed to update team lead"))
		return
	}

	c.Status(http.StatusNoContent)
}
