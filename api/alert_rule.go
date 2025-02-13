package api

import (
	"ipoc-golang-api/global"
	"ipoc-golang-api/models"
	"ipoc-golang-api/services/alerts"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm/clause"
)

// @Summary Get All Alert Rule
// @Description only admin
// @Tags AlertRule
// @Accept  json
// @Produce  json
// @Success 200 {object} []models.AlertRuleRes
// @Security ApiKeyAuth
// @Router /alert-rule/get-all [get]
func (h *AlertHandler) GetAllAlertRule(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	res := alerts.GetAllAlertRules(user.Realm)
	c.JSON(http.StatusOK, res)
}

// @Summary Get Alert Rule By ID
// @Description only admin
// @Tags AlertRule
// @Accept  json
// @Produce  json
// @Success 200 {object} models.AlertRuleRes
// @Security ApiKeyAuth
// @Router /alert-rule/{id} [get]
func (h *AlertHandler) GetAlertRuleByID(c *gin.Context) {
	//user := c.Keys["user"].(models.SSOUser)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadGateway, err.Error())
		return
	}
	res := alerts.GetAlertRuleResByID(id)
	c.JSON(http.StatusOK, res)
}

// @Summary Copy Alert Rule By ID
// @Description only admin
// @Tags AlertRule
// @Accept  json
// @Produce  json
// @Success 200 {object} models.AlertRuleRes
// @Security ApiKeyAuth
// @Router /copy-alert-rule/{id} [get]
func (h *AlertHandler) CopyAlertRuleByID(c *gin.Context) {
	//user := c.Keys["user"].(models.SSOUser)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadGateway, err.Error())
		return
	}
	newRule := alerts.GetAlertRuleResByID(id)

	newRule.ID = 0
	newRule.Name = newRule.Name + " Copy"
	newRule.Enabled = false

	//* Create DB
	res, err := alerts.CreateAlertRule(newRule)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	//* res
	res = alerts.GetAlertRuleResByID(res.ID)
	c.JSON(http.StatusOK, res)
}

// @Summary Create AlertRule
// @Tags AlertRule
// @Accept  json
// @Produce  json
// @Param rule body models.AlertRuleRes true "rule"
// @Success 200 {object} models.AlertRuleRes
// @Security ApiKeyAuth
// @Router /alert-rule/create [post]
func (h *AlertHandler) CreateAlertRule(c *gin.Context) {
	user := c.Keys["user"].(models.SSOUser)
	body := new(models.AlertRuleRes)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	body.RealmName = user.Realm

	//* check Realm name & rule name 不能重複
	check := alerts.CheckAlertRuleName(body.RealmName, body.ID, body.Name)
	if !check.Success {
		c.JSON(http.StatusBadRequest, check.Msg)
		return
	}

	//* Create DB
	res, err := alerts.CreateAlertRule(*body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	//* res
	res = alerts.GetAlertRuleResByID(res.ID)
	c.JSON(http.StatusOK, res)
}

// @Summary Test Alert Rule
// @Tags AlertRule
// @Accept  json
// @Produce  json
// @Param rule body models.AlertRuleRes true "rule"
// @Success 200 {object}  models.Response
// @Security ApiKeyAuth
// @Router /alert-rule/test [post]
func (h *AlertHandler) TestAlertRule(c *gin.Context) {
	res := models.Response{}
	test_rule := new(models.AlertRuleRes)

	err := c.Bind(&test_rule)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	metric_rule, err := alerts.GetMetricRuleByID(test_rule.MetricRule.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	//* models.AlertRuleRes 轉 models.AlertRule
	db_rule := models.AlertRule{
		RealmName: c.Keys["user"].(models.SSOUser).Realm,
		ID:        0,
		Name:      test_rule.Name + " Test",

		DataRange: test_rule.DataRange,

		SilencePeriod:      test_rule.SilencePeriod,
		Enabled:            false,
		ResourceGroupID:    test_rule.ResourceGroup.ID,
		ResourceGroup:      test_rule.ResourceGroup,
		MetricRuleID:       test_rule.MetricRule.ID,
		MetricRule:         metric_rule,
		AlertMuteID:        0,
		AlertingTemplateID: metric_rule.AlertingTemplateID,
		StateTemplateID:    metric_rule.StateTemplateID,
		AlertingTemplate:   test_rule.AlertingTemplate,
		StateTemplate:      test_rule.StateTemplate,
		Times:              test_rule.Times,
		MaxAlert:           test_rule.MaxAlert,

		InfoThreshold: test_rule.InfoThreshold,
		WarnThreshold: test_rule.WarnThreshold,
		CritThreshold: test_rule.CritThreshold,

		InfoDuration: test_rule.InfoDuration,
		WarnDuration: test_rule.WarnDuration,
		CritDuration: test_rule.CritDuration,

		ActiveTime:       test_rule.ActiveTime,
		ActiveWeekday:    test_rule.ActiveWeekday,
		AlertContacts:    []models.AlertContact{},
		AlertRuleDetails: []models.AlertRuleDetail{},
		AlertLabels:      test_rule.AlertLabels,
	}
	db := global.Mysql
	for _, a := range test_rule.AlertContacts {

		var db_contact models.AlertContact

		err := db.Preload(clause.Associations).Where(&models.AlertContact{ID: a.ID}).Find(&db_contact).Error
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}

		// * find levels
		var levels []string
		err = db.Model(&models.AlertContactLevel{}).Where(&models.AlertContactLevel{
			AlertContactID: a.ID,
		}).Distinct().Pluck("level", &levels).Error
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}
		db_contact.Levels = levels

		for _, dd := range db_contact.AlertContactDetails {
			var d models.AlertContactDetail
			d.Key = dd.Key
			d.Value = dd.Value
			db_contact.AlertContactDetails = append(db_contact.AlertContactDetails, d)
		}
		db_rule.AlertContacts = append(db_rule.AlertContacts, db_contact)
	}

	err = db.Preload(clause.Associations).Where(models.ResourceGroup{
		ID: db_rule.ResourceGroup.ID,
	}).Find(&db_rule.ResourceGroup).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err = db.Where(models.AlertTemplate{
		ID: metric_rule.AlertingTemplateID,
	}).Find(&db_rule.AlertingTemplate).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	err = db.Where(models.AlertTemplate{
		ID: metric_rule.StateTemplateID,
	}).Find(&db_rule.StateTemplate).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	db_rule.AlertRuleDetails = alerts.ResourcePartitionsFormatAlertRuleDetails(*test_rule)

	//* Alert Label
	for i, label := range db_rule.AlertLabels {
		var db_label models.AlertLabel
		err := db.Where(&models.AlertLabel{
			ID: label.ID,
		}).Find(&db_label).Error
		if err != nil {
			c.JSON(http.StatusBadRequest, err.Error())
			return
		}
		db_rule.AlertLabels[i] = db_label

	}

	alerts.RunAlertRuleByResoureGroup(db_rule)
	res.Success = true
	c.JSON(http.StatusOK, res)
}

// @Summary Update Alert Rule
// @Description only admin
// @Tags AlertRule
// @Accept  json
// @Produce  json
// @Param group body models.AlertRuleRes true "update"
// @Success 200 {object} models.AlertRuleRes
// @Security ApiKeyAuth
// @Router /alert-rule/update [put]
func (h *AlertHandler) UpdateAlertRule(c *gin.Context) {
	body := new(models.AlertRuleRes)

	err := c.Bind(&body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	body.RealmName = c.Keys["user"].(models.SSOUser).Realm

	check := alerts.CheckAlertRuleName(body.RealmName, body.ID, body.Name)
	if !check.Success {
		c.JSON(http.StatusBadRequest, check.Msg)
		return
	}

	// Update DB
	res, err := alerts.UpdateAlertRule(*body)
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}

	c.JSON(http.StatusOK, res)
}

// @Summary Delete Alert Rule By ID
// @Description only admin
// @Tags AlertRule
// @Accept  json
// @Produce  json
// @Param id path int true "id"
// @Success 200 {object} models.Response
// @Security ApiKeyAuth
// @Router /alert-rule/delete/{id} [delete]
func (h *AlertHandler) DeleteAlertRuleByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadGateway, err.Error())
		return
	}

	// Delete DB
	res := alerts.DeleteAlertRuleByID(id)
	if !res.Success {
		c.JSON(http.StatusBadRequest, res.Msg)
		return
	}
	c.JSON(http.StatusOK, res.Msg)
}
