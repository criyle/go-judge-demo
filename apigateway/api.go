package main

import (
	"fmt"
	"io"
	"net/http"

	"github.com/criyle/go-judger-demo/pb"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/jsonpb"
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
	resp, err := a.client.Submission(c, &pb.SubmissionRequest{
		Id: id,
	})
	if err != nil {
		c.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	c.JSON(http.StatusOK, resp.Submissions)
}

func (a *api) apiSubmit(c *gin.Context) {
	// limit upload size
	if c.Request.ContentLength > maxLimit {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			fmt.Sprintf("Upload size too large: %d", c.Request.ContentLength))
		return
	}
	body := &io.LimitedReader{R: c.Request.Body, N: maxLimit}
	var req pb.SubmitRequest
	if err := jsonpb.Unmarshal(body, &req); err != nil {
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
