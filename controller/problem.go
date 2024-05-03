package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"online-judge/pkg"
	"online-judge/pkg/resp"
	"online-judge/services"
	"strconv"
)

// GetProblemList 获取题目列表接口
// @Summary 获取题目列表
// @Description 获取题目列表接口
// @Success 200 {object} _Response "获取题目列表成功"
// @Failure 200 {object} _Response "服务器内部错误"
// @Router /problem-list [GET]
func GetProblemList(c *gin.Context) {
	var getProblemList services.Problem
	data, err := getProblemList.GetProblemList()
	if err != nil {
		resp.ResponseError(c, resp.CodeInternalServerError)
		return
	}
	resp.ResponseSuccess(c, data)
}

// GetProblemDetail 获取单个题目详细接口
// @Summary 获取单个题目详细
// @Description 获取单个题目详细接口
// @Accept multipart/form-data
// @Produce json,multipart/form-data
// @Param problem_id query string true "题目ID"
// @Success 200 {object} _Response "获取成功"
// @Failure 200 {object} _Response "题目ID不存在"
// @Failure 200 {object} _Response "服务器内部错误"
// @Router /problem/{problem_id} [GET]
func GetProblemDetail(c *gin.Context) {
	var getProblemDetail services.Problem
	pid := c.Query("problem_id")
	getProblemDetail.ProblemID = pid

	data, err := getProblemDetail.GetProblemDetail()
	if err != nil {
		resp.ResponseError(c, resp.CodeProblemIDNotExist)
		return
	}
	resp.ResponseSuccess(c, data)
}

// CreateProblem 创建新题目接口
// @Summary 创建新题目
// @Description 创建新题目接口
// @Accept multipart/form-data
// @Produce json,multipart/form-data
// @Param title formData string true "题目标题"
// @Param content formData string true "题目内容"
// @Param difficulty formData string true "题目难度"
// @Param max_runtime formData int true "时间限制"
// @Param max_memory formData int true "内存限制"
// @Param test_cases formData []string true "测试样例集" collectionFormat(multi)
// @Success 200 {object} _Response "创建成功"
// @Failure 200 {object} _Response "参数错误"
// @Failure 200 {object} _Response "服务器内部错误"
// @Router /problem-create [POST]
func CreateProblem(c *gin.Context) {
	var createProblem services.Problem

	title := c.PostForm("title")
	content := c.PostForm("content")
	difficulty := c.PostForm("difficulty")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMemory, _ := strconv.Atoi(c.PostForm("max_memory"))

	testCase := c.PostFormArray("test_cases")
	if len(testCase) == 0 {
		zap.L().Error("testCase is empty")
		resp.ResponseError(c, resp.CodeInvalidParam)
		return
	}
	//fmt.Println(title)
	//fmt.Println(content)
	//fmt.Println(difficulty)
	//fmt.Println(maxRuntime)
	//fmt.Println(maxMemory)
	//fmt.Println(testCase)
	createProblem.ProblemID = pkg.GetUUID()
	createProblem.Content = content
	createProblem.Difficulty = difficulty
	createProblem.Title = title
	createProblem.MaxRuntime = maxRuntime
	createProblem.MaxMemory = maxMemory

	tCase := make([]*services.TestCase, 0)
	for _, value := range testCase {
		caseMap := make(map[string]string)
		err := json.Unmarshal([]byte(value), &caseMap)
		// 检测Map某个键是否存在
		_, iok := caseMap["input"]
		_, ook := caseMap["expected"]
		if err != nil || !iok || !ook {
			resp.ResponseError(c, resp.CodeTestCaseFormatError)
			if err != nil {
				zap.L().Error("caseMap unmarshal error ", zap.Error(err))
			}
			return
		}
		tCase = append(tCase, &services.TestCase{
			TID:      pkg.GetUUID(),
			PID:      createProblem.ProblemID,
			Input:    caseMap["input"],
			Expected: caseMap["expected"],
		})
	}
	createProblem.TestCases = tCase
	response := createProblem.CreateProblem()
	switch response.Code {
	case resp.Success:
		resp.ResponseSuccess(c, resp.CodeSuccess)

	case resp.ProblemAlreadyExist:
		resp.ResponseError(c, resp.CodeProblemExist)

	case resp.CreateProblemError:
		resp.ResponseError(c, resp.CodeInternalServerError)

	default:
		resp.ResponseError(c, resp.CodeInternalServerError)
	}

}

// UpdateProblem 更新题目信息接口
// @Summary 更新题目信息
// @Description 更新题目信息接口
// @Accept multipart/form-data
// @Produce json
// @Param problem_id query string true "题目ID"
// @Param title query string false "题目标题"
// @Param content query string false "题目内容"
// @Param difficulty query string false "题目难度"
// @Param max_runtime query string false "时间限制"
// @Param max_memory query string false "内存限制"
// @Param max_memory query string false "内存限制"
// @Param test_cases formData []string false "测试样例集" collectionFormat(multi)
// @Success 200 {object} _Response "修改成功"
// @Failure 200 {object} _Response "题目ID不存在"
// @Failure 200 {object} _Response "服务器内部错误"
// @Router /problem/{problem_id} [PUT]
func UpdateProblem(c *gin.Context) {
	var updateProblem services.Problem

	title := c.PostForm("title")
	content := c.PostForm("content")
	difficulty := c.PostForm("difficulty")
	maxRuntime, _ := strconv.Atoi(c.PostForm("max_runtime"))
	maxMemory, _ := strconv.Atoi(c.PostForm("max_memory"))
	testCase := c.PostFormArray("test_cases")

	updateProblem.UpdateProblem()

}

// DeleteProblem 删除题目
// @Summary 更新题目信息
// @Description 更新题目信息接口
// @Accept multipart/form-data
// @Produce json
// @Param problem_id query string true "题目ID"
// @Success 200 {object} _Response "删除成功"
// @Failure 200 {object} _Response "题目ID不存在"
// @Failure 200 {object} _Response "服务器内部错误"
// @Router /problem/{problem_id} [DELETE]
func DeleteProblem(c *gin.Context) {
	var deleteProblem services.Problem
	deleteProblem.DeleteProblem()

}
