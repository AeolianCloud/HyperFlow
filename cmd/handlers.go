package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"hyperflow/internal/operations"
	"hyperflow/internal/pve"
)

// ErrorDetail 标准错误详情，遵循微软 REST API Guidelines
type ErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ErrorResponse 标准错误响应体
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// OperationResponse 异步操作状态响应，遵循微软 LRO 规范
type OperationResponse struct {
	ID               string              `json:"id"`
	Status           string              `json:"status"`
	ResourceLocation string              `json:"resourceLocation,omitempty"`
	Error            *OperationErrorBody `json:"error,omitempty"`
}

// OperationErrorBody LRO 失败时的错误体
type OperationErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// respondError 输出符合微软 REST API Guidelines 的标准错误响应
func respondError(c *gin.Context, status int, code, message string) {
	c.JSON(status, ErrorResponse{Error: ErrorDetail{Code: code, Message: message}})
}

// respondOK 输出 200 响应，直接输出资源，不包装 data 字段
func respondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, data)
}

func handlePveError(c *gin.Context, err error) {
	if pveErr, ok := err.(*pve.PveError); ok {
		switch pveErr.StatusCode {
		case http.StatusNotFound:
			respondError(c, http.StatusNotFound, "NotFound", pveErr.Message)
		case http.StatusConflict:
			respondError(c, http.StatusConflict, "Conflict", pveErr.Message)
		case 502:
			respondError(c, http.StatusBadGateway, "BadGateway", "PVE server unreachable: "+pveErr.Message)
		default:
			respondError(c, pveErr.StatusCode, "InternalServerError", pveErr.Message)
		}
		return
	}
	respondError(c, http.StatusInternalServerError, "InternalServerError", err.Error())
}

var nodesSvcGlobal *pve.NodesService
var vmsSvcGlobal *pve.VmsService
var storageSvcGlobal *pve.StorageService
var operationsSvcGlobal *operations.Service

// listNodes godoc
// @Summary      列出所有节点
// @Description  返回 PVE 集群中所有节点列表
// @Tags         nodes
// @Produce      json
// @Success      200  {array}   map[string]any
// @Failure      500  {object}  ErrorResponse
// @Failure      502  {object}  ErrorResponse
// @Router       /nodes [get]
func listNodes(c *gin.Context) {
	nodes, err := nodesSvcGlobal.ListNodes(requestContextFromGin(c))
	if err != nil {
		handlePveError(c, err)
		return
	}
	respondOK(c, nodes)
}

// getNode godoc
// @Summary      获取节点详情
// @Description  返回指定节点的详细状态信息
// @Tags         nodes
// @Produce      json
// @Param        node  path      string  true  "节点名称"
// @Success      200   {object}  map[string]any
// @Failure      404   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Failure      502   {object}  ErrorResponse
// @Router       /nodes/{node} [get]
func getNode(c *gin.Context) {
	node := c.Param("node")
	data, err := nodesSvcGlobal.GetNode(requestContextFromGin(c), node)
	if err != nil {
		handlePveError(c, err)
		return
	}
	var result any
	json.Unmarshal(data, &result)
	respondOK(c, result)
}

// listVms godoc
// @Summary      列出节点上所有虚拟机
// @Description  返回指定节点上所有 QEMU 虚拟机列表
// @Tags         vms
// @Produce      json
// @Param        node  path      string  true  "节点名称"
// @Success      200   {array}   pve.VM
// @Failure      500   {object}  ErrorResponse
// @Failure      502   {object}  ErrorResponse
// @Router       /nodes/{node}/vms [get]
func listVms(c *gin.Context) {
	node := c.Param("node")
	vms, err := vmsSvcGlobal.ListVms(requestContextFromGin(c), node)
	if err != nil {
		handlePveError(c, err)
		return
	}
	respondOK(c, vms)
}

// getVm godoc
// @Summary      获取指定虚拟机信息
// @Description  返回指定虚拟机的当前状态信息
// @Tags         vms
// @Produce      json
// @Param        node  path      string  true  "节点名称"
// @Param        vmid  path      string  true  "虚拟机 ID"
// @Success      200   {object}  map[string]any
// @Failure      404   {object}  ErrorResponse
// @Failure      500   {object}  ErrorResponse
// @Failure      502   {object}  ErrorResponse
// @Router       /nodes/{node}/vms/{vmid} [get]
func getVm(c *gin.Context) {
	node := c.Param("node")
	vmid := c.Param("vmid")
	data, err := vmsSvcGlobal.GetVm(requestContextFromGin(c), node, vmid)
	if err != nil {
		handlePveError(c, err)
		return
	}
	var result any
	json.Unmarshal(data, &result)
	respondOK(c, result)
}

// startVm godoc
// @Summary      启动虚拟机
// @Description  异步启动指定虚拟机，返回 LRO Operation-Location header
// @Tags         vms
// @Produce      json
// @Param        node  path  string  true  "节点名称"
// @Param        vmid  path  string  true  "虚拟机 ID"
// @Success      202
// @Header       202  {string}  Operation-Location  "/api/pve/operations/{id}"
// @Failure      404  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Failure      502  {object}  ErrorResponse
// @Router       /nodes/{node}/vms/{vmid}/start [post]
func startVm(c *gin.Context) {
	node := c.Param("node")
	vmid := c.Param("vmid")
	ctx := requestContextFromGin(c)
	data, err := vmsSvcGlobal.StartVm(ctx, node, vmid)
	if err != nil {
		handlePveError(c, err)
		return
	}
	upid, err := decodeUPID(data)
	if err != nil {
		respondError(c, http.StatusBadGateway, "BadGateway", "invalid PVE task response: "+err.Error())
		return
	}
	op, err := operationsSvcGlobal.CreateOperation(ctx, node, upid, "/api/pve/nodes/"+node+"/vms/"+vmid)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "InternalServerError", err.Error())
		return
	}
	c.Header("Operation-Location", "/api/pve/operations/"+op.ID)
	c.Status(http.StatusAccepted)
}

// stopVm godoc
// @Summary      停止虚拟机
// @Description  异步停止指定虚拟机，返回 LRO Operation-Location header
// @Tags         vms
// @Produce      json
// @Param        node  path  string  true  "节点名称"
// @Param        vmid  path  string  true  "虚拟机 ID"
// @Success      202
// @Header       202  {string}  Operation-Location  "/api/pve/operations/{id}"
// @Failure      404  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Failure      502  {object}  ErrorResponse
// @Router       /nodes/{node}/vms/{vmid}/stop [post]
func stopVm(c *gin.Context) {
	node := c.Param("node")
	vmid := c.Param("vmid")
	ctx := requestContextFromGin(c)
	data, err := vmsSvcGlobal.StopVm(ctx, node, vmid)
	if err != nil {
		handlePveError(c, err)
		return
	}
	upid, err := decodeUPID(data)
	if err != nil {
		respondError(c, http.StatusBadGateway, "BadGateway", "invalid PVE task response: "+err.Error())
		return
	}
	op, err := operationsSvcGlobal.CreateOperation(ctx, node, upid, "/api/pve/nodes/"+node+"/vms/"+vmid)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "InternalServerError", err.Error())
		return
	}
	c.Header("Operation-Location", "/api/pve/operations/"+op.ID)
	c.Status(http.StatusAccepted)
}

// deleteVm godoc
// @Summary      删除虚拟机
// @Description  异步删除指定虚拟机（虚拟机须处于停止状态），返回 LRO Operation-Location header
// @Tags         vms
// @Produce      json
// @Param        node  path  string  true  "节点名称"
// @Param        vmid  path  string  true  "虚拟机 ID"
// @Success      202
// @Header       202  {string}  Operation-Location  "/api/pve/operations/{id}"
// @Failure      404  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Failure      502  {object}  ErrorResponse
// @Router       /nodes/{node}/vms/{vmid} [delete]
func deleteVm(c *gin.Context) {
	node := c.Param("node")
	vmid := c.Param("vmid")
	ctx := requestContextFromGin(c)
	data, err := vmsSvcGlobal.DeleteVm(ctx, node, vmid)
	if err != nil {
		handlePveError(c, err)
		return
	}
	upid, err := decodeUPID(data)
	if err != nil {
		respondError(c, http.StatusBadGateway, "BadGateway", "invalid PVE task response: "+err.Error())
		return
	}
	op, err := operationsSvcGlobal.CreateOperation(ctx, node, upid, "")
	if err != nil {
		respondError(c, http.StatusInternalServerError, "InternalServerError", err.Error())
		return
	}
	c.Header("Operation-Location", "/api/pve/operations/"+op.ID)
	c.Status(http.StatusAccepted)
}

// createVm godoc
// @Summary      新建虚拟机并导入磁盘
// @Description  通过 PVE 创建新虚拟机，并在创建时通过 import-from 导入指定磁盘卷。
// @Description  支持可选的 CloudInit 配置（ciUser、ciPassword、sshKeys、ipConfig0、nameserver、searchDomain）；
// @Description  当请求体包含任意 CloudInit 字段时，系统自动附加 CloudInit 驱动盘（ide2）及对应配置。
// @Description  ciPackages 非空时在 snippetsStorage 中生成 cloud-init user-data Snippet（自动执行 package_update/upgrade，并将 ciUser、ciPassword、sshKeys 写入 user-data）并通过 cicustom 引用（snippetsStorage 此时必填）。
// @Description  成功后响应 Operation-Location 及 Location 头。
// @Tags         vms
// @Accept       json
// @Produce      json
// @Param        node  path  string               true  "节点名称"
// @Param        body  body  pve.CreateVmRequest  true  "创建参数（vmid、cores、memory、diskSource、storage 为必填；name 为可选，不填时自动生成随机主机名；CloudInit 字段均为可选；ciPackages 非空时 snippetsStorage 必填）"
// @Success      202
// @Header       202  {string}  Operation-Location  "/api/pve/operations/{id}"
// @Header       202  {string}  Location            "/api/pve/nodes/{node}/vms/{vmid}"
// @Failure      400  {object}  ErrorResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      409  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Failure      502  {object}  ErrorResponse
// @Router       /nodes/{node}/vms [post]
func createVm(c *gin.Context) {
	node := c.Param("node")
	ctx := requestContextFromGin(c)
	var req pve.CreateVmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "BadRequest", "invalid request body: "+err.Error())
		return
	}
	if req.VMID == 0 || req.Cores == 0 || req.Memory == 0 || req.DiskSource == "" || req.Storage == "" {
		respondError(c, http.StatusBadRequest, "BadRequest", "vmid, cores, memory, diskSource and storage are required")
		return
	}
	if len(req.CIPackages) > 0 && req.SnippetsStorage == "" {
		respondError(c, http.StatusBadRequest, "BadRequest", "snippetsStorage is required when ciPackages is specified")
		return
	}
	data, err := vmsSvcGlobal.CreateVm(ctx, node, req)
	if err != nil {
		handlePveError(c, err)
		return
	}
	upid, err := decodeUPID(data)
	if err != nil {
		respondError(c, http.StatusBadGateway, "BadGateway", "invalid PVE task response: "+err.Error())
		return
	}
	vmLocation := "/api/pve/nodes/" + node + "/vms/" + fmt.Sprint(req.VMID)
	op, err := operationsSvcGlobal.CreateOperation(ctx, node, upid, vmLocation)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "InternalServerError", err.Error())
		return
	}
	c.Header("Operation-Location", "/api/pve/operations/"+op.ID)
	c.Header("Location", vmLocation)
	c.Status(http.StatusAccepted)
}

// listStorage godoc
// @Summary      列出所有存储
// @Description  返回 PVE 中所有存储资源列表
// @Tags         storage
// @Produce      json
// @Success      200  {array}   map[string]any
// @Failure      500  {object}  ErrorResponse
// @Failure      502  {object}  ErrorResponse
// @Router       /storage [get]
func listStorage(c *gin.Context) {
	storages, err := storageSvcGlobal.ListStorage(requestContextFromGin(c))
	if err != nil {
		handlePveError(c, err)
		return
	}
	respondOK(c, storages)
}

// getOperation godoc
// @Summary      获取异步操作状态
// @Description  返回指定异步操作的当前状态，遵循 Microsoft REST API Guidelines LRO 模式。
// @Tags         operations
// @Param        id  path  string  true  "操作 ID"
// @Produce      json
// @Success      200  {object}  OperationResponse
// @Failure      404  {object}  ErrorResponse
// @Failure      500  {object}  ErrorResponse
// @Router       /operations/{id} [get]
func getOperation(c *gin.Context) {
	id := c.Param("id")
	op, err := operationsSvcGlobal.GetOperation(requestContextFromGin(c), id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "InternalServerError", err.Error())
		return
	}
	if op == nil {
		respondError(c, http.StatusNotFound, "NotFound", "operation not found")
		return
	}
	respondOK(c, buildOperationResponse(op))
}

func buildOperationResponse(op *operations.Operation) OperationResponse {
	resp := OperationResponse{
		ID:               op.ID,
		Status:           op.Status,
		ResourceLocation: op.ResourceLocation,
	}
	if op.ErrorCode != "" {
		resp.Error = &OperationErrorBody{Code: op.ErrorCode, Message: op.ErrorMessage}
	}
	return resp
}

func decodeUPID(data []byte) (string, error) {
	var upid string
	if err := json.Unmarshal(data, &upid); err != nil {
		return "", err
	}
	if upid == "" {
		return "", fmt.Errorf("empty task id")
	}
	return upid, nil
}

// registerNodesRoutes 注册节点路由
func registerNodesRoutes(rg *gin.RouterGroup, svc *pve.NodesService) {
	nodesSvcGlobal = svc
	rg.GET("", listNodes)
	rg.GET("/:node", getNode)
}

// registerVmsRoutes 注册虚拟机路由
func registerVmsRoutes(rg *gin.RouterGroup, svc *pve.VmsService) {
	vmsSvcGlobal = svc
	rg.GET("", listVms)
	rg.POST("", createVm)
	rg.GET("/:vmid", getVm)
	rg.POST("/:vmid/start", startVm)
	rg.POST("/:vmid/stop", stopVm)
	rg.DELETE("/:vmid", deleteVm)
}

// registerStorageRoutes 注册存储路由
func registerStorageRoutes(rg *gin.RouterGroup, svc *pve.StorageService) {
	storageSvcGlobal = svc
	rg.GET("", listStorage)
}

// registerOperationsRoutes 注册操作路由
func registerOperationsRoutes(rg *gin.RouterGroup, svc *operations.Service) {
	operationsSvcGlobal = svc
	rg.GET("/:id", getOperation)
}
