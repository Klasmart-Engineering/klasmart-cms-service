package api

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"gitlab.badanamu.com.cn/calmisland/kidsloop2/utils"
	"net/http"
)

type TokenResponse struct {
	Token string `json:"token"`
}
type SignatureResponse struct {
	URL string `json:"url"`
}

// @Summary h5pSignature
// @ID h5pSignature
// @Description signature url for h5p
// @Accept json
// @Produce json
// @Param url query string false "url to signature"
// @Tags crypto
// @Success 200 {object} SignatureResponse
// @Failure 500 {object} StatusInternalServerError
// @Router /crypto/h5p/signature [get]
func (s Server) h5pSignature(c *gin.Context) {
	operator := GetOperator(c)
	urlStr := c.Query("url")
	res, err := utils.URLSignature(operator.UserID, urlStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}
	h5pPath := fmt.Sprintf("%v?badanamuId=%v&timestamp=%016x&randNum=%016x&signature=%v", urlStr, operator.UserID, res.Timestamp, res.RandNum, res.Signature)

	c.JSON(http.StatusOK, SignatureResponse{URL: h5pPath})
}

// @Summary generateH5pJWT
// @ID generateH5pJWT
// @Description generate JWT for h5p
// @Accept json
// @Produce json
// @Param sub query string false "subject for jwt"
// @Param content_id query string false "content id to operate"
// @Tags crypto
// @Success 200 {object} TokenResponse
// @Failure 500 {object} StatusInternalServerError
// @Router /crypto/h5p/jwt [get]
func (s Server) generateH5pJWT(c *gin.Context) {
	sub := c.Query("sub")
	contentId := c.Query("content_id")
	token, err := utils.GenerateH5pJWT(c.Request.Context(), sub, contentId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responseMsg(err.Error()))
		return
	}

	c.JSON(http.StatusOK, TokenResponse{Token: token})
}