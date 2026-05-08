package web

import (
	"errors"
	"strconv"

	"moebot-next/internal/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

func (s *Server) handleListFilterTemplates(c *fiber.Ctx) error {
	ts, err := s.DB.ListFilterTemplates()
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	out := make([]filterTemplatePayload, 0, len(ts))
	for i := range ts {
		usage, _ := s.DB.CountFilterAppsByTemplate(ts[i].ID)
		out = append(out, templateToPayload(&ts[i], usage))
	}
	return c.JSON(fiber.Map{"items": out})
}

func (s *Server) handleGetFilterTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	t, err := s.DB.GetFilterTemplate(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "template not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	usage, _ := s.DB.CountFilterAppsByTemplate(t.ID)
	return c.JSON(templateToPayload(t, usage))
}

func (s *Server) handleCreateFilterTemplate(c *fiber.Ctx) error {
	var p filterTemplatePayload
	if err := c.BodyParser(&p); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if p.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name is required")
	}
	if existing, err := s.DB.GetFilterTemplateByName(p.Name); err == nil && existing != nil {
		return fiber.NewError(fiber.StatusConflict, "template name already in use")
	} else if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	t := &models.FilterTemplate{}
	payloadToTemplate(&p, t)
	if err := s.DB.CreateFilterTemplate(t); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if s.Filter != nil {
		_ = s.Filter.Reload(c.UserContext())
	}
	return c.JSON(templateToPayload(t, 0))
}

func (s *Server) handleUpdateFilterTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var p filterTemplatePayload
	if err := c.BodyParser(&p); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	t, err := s.DB.GetFilterTemplate(uint(id))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "template not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	payloadToTemplate(&p, t)
	if err := s.DB.UpdateFilterTemplate(t); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	if s.Filter != nil {
		_ = s.Filter.Reload(c.UserContext())
	}
	usage, _ := s.DB.CountFilterAppsByTemplate(t.ID)
	return c.JSON(templateToPayload(t, usage))
}

func (s *Server) handleDeleteFilterTemplate(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	if err := s.DB.DeleteFilterTemplate(uint(id)); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "template not found")
		}
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	if s.Filter != nil {
		_ = s.Filter.Reload(c.UserContext())
	}
	return c.JSON(fiber.Map{"ok": true})
}
