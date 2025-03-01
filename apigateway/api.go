package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/criyle/go-judge-demo/pb"
	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
)

const maxLimit = 64 << 10 // 64k

type api struct {
	client pb.DemoBackendClient
}

func (a *api) Register(r *gin.RouterGroup) {
	r.GET("/submission", a.apiSubmission)
	r.POST("/submit", a.apiSubmit)
}

func (a *api) apiSubmission(c *gin.Context) {
	id := c.Query("id")
	resp, err := a.client.Submission(c, pb.SubmissionRequest_builder{
		Id: &id,
	}.Build())
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	ct, err := protojson.Marshal(resp)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.Data(http.StatusOK, "application/json; charset=utf-8", ct)
}

func (a *api) apiSubmit(c *gin.Context) {
	// limit upload size
	if c.Request.ContentLength > maxLimit {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			fmt.Sprintf("Upload size too large: %d", c.Request.ContentLength))
		return
	}
	body, err := io.ReadAll(io.LimitReader(c.Request.Body, maxLimit))
	if err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	var req pb.SubmitRequest
	if err := protojson.Unmarshal(body, &req); err != nil {
		c.AbortWithError(http.StatusBadRequest, err)
		return
	}
	resp, err := a.client.Submit(c, &req)
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, resp)
}
