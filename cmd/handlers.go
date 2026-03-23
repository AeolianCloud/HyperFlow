package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"hyperflow/internal/pve"
)

func respondOK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func respondAccepted(c *gin.Context, data any) {
	c.JSON(http.StatusAccepted, gin.H{"data": data})
}

func handlePveError(c *gin.Context, err error) {
	if pveErr, ok := err.(*pve.PveError); ok {
		switch pveErr.StatusCode {
		case http.StatusNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": pveErr.Message})
		case http.StatusConflict:
			c.JSON(http.StatusConflict, gin.H{"error": pveErr.Message})
		case 502:
			c.JSON(http.StatusBadGateway, gin.H{"error": "PVE server unreachable: " + pveErr.Message})
		default:
			c.JSON(pveErr.StatusCode, gin.H{"error": pveErr.Message})
		}
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
}

var nodesSvcGlobal *pve.NodesService
var vmsSvcGlobal *pve.VmsService
var storageSvcGlobal *pve.StorageService

// listNodes godoc
// @Summary      列出所有节点
// @Description  返回 PVE 集群中所有节点列表
// @Tags         nodes
// @Produce      json
// @Success      200  {object}  map[string]any
// @Failure      500  {object}  map[string]string
// @Failure      502  {object}  map[string]string
// @Router       /nodes [get]
func listNodes(c *gin.Context) {
	nodes, err := nodesSvcGlobal.ListNodes()
	if err != nil {
		handlePveError(c, err)
		return
	}
	respondOK(c, nodes)
}

// getNode godoc
// @Summary      获取指定节点信息
// @Description  返回指定节点的详细状态信息
// @Tags         nodes
// @Produce      json
// @Param        node  path      string  true  "节点名称"
// @Success      200   {object}  map[string]any
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Failure      502   {object}  map[string]string
// @Router       /nodes/{node} [get]
func getNode(c *gin.Context) {
	node := c.Param("node")
	data, err := nodesSvcGlobal.GetNode(node)
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
// @Success      200   {object}  map[string]any
// @Failure      500   {object}  map[string]string
// @Failure      502   {object}  map[string]string
// @Router       /nodes/{node}/vms [get]
func listVms(c *gin.Context) {
	node := c.Param("node")
	vms, err := vmsSvcGlobal.ListVms(node)
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
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Failure      502   {object}  map[string]string
// @Router       /nodes/{node}/vms/{vmid} [get]
func getVm(c *gin.Context) {
	node := c.Param("node")
	vmid := c.Param("vmid")
	data, err := vmsSvcGlobal.GetVm(node, vmid)
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
// @Description  异步启动指定虚拟机，返回 PVE 任务 ID
// @Tags         vms
// @Produce      json
// @Param        node  path      string  true  "节点名称"
// @Param        vmid  path      string  true  "虚拟机 ID"
// @Success      202   {object}  map[string]any
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Failure      502   {object}  map[string]string
// @Router       /nodes/{node}/vms/{vmid}/start [post]
func startVm(c *gin.Context) {
	node := c.Param("node")
	vmid := c.Param("vmid")
	data, err := vmsSvcGlobal.StartVm(node, vmid)
	if err != nil {
		handlePveError(c, err)
		return
	}
	var result any
	json.Unmarshal(data, &result)
	respondAccepted(c, result)
}

// stopVm godoc
// @Summary      停止虚拟机
// @Description  异步停止指定虚拟机，返回 PVE 任务 ID
// @Tags         vms
// @Produce      json
// @Param        node  path      string  true  "节点名称"
// @Param        vmid  path      string  true  "虚拟机 ID"
// @Success      202   {object}  map[string]any
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Failure      502   {object}  map[string]string
// @Router       /nodes/{node}/vms/{vmid}/stop [post]
func stopVm(c *gin.Context) {
	node := c.Param("node")
	vmid := c.Param("vmid")
	data, err := vmsSvcGlobal.StopVm(node, vmid)
	if err != nil {
		handlePveError(c, err)
		return
	}
	var result any
	json.Unmarshal(data, &result)
	respondAccepted(c, result)
}

// deleteVm godoc
// @Summary      删除虚拟机
// @Description  异步删除指定虚拟机，返回 PVE 任务 ID
// @Tags         vms
// @Produce      json
// @Param        node  path      string  true  "节点名称"
// @Param        vmid  path      string  true  "虚拟机 ID"
// @Success      202   {object}  map[string]any
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Failure      502   {object}  map[string]string
// @Router       /nodes/{node}/vms/{vmid} [delete]
func deleteVm(c *gin.Context) {
	node := c.Param("node")
	vmid := c.Param("vmid")
	data, err := vmsSvcGlobal.DeleteVm(node, vmid)
	if err != nil {
		handlePveError(c, err)
		return
	}
	var result any
	json.Unmarshal(data, &result)
	respondAccepted(c, result)
}

// createVm godoc
// @Summary      新建虚拟机并导入磁盘
// @Description  通过 PVE 创建新虚拟机，并在创建时通过 import-from 导入指定磁盘卷。
// @Description  支持可选的 CloudInit 配置（ciUser、ciPassword、sshKeys、ipConfig0、nameserver、searchDomain）；
// @Description  当请求体包含任意 CloudInit 字段时，系统自动附加 CloudInit 驱动盘（ide2）及对应配置。
// @Description  成功后响应 Location 头指向新虚拟机资源路径。
// @Tags         vms
// @Accept       json
// @Produce      json
// @Param        node  path      string               true  "节点名称"
// @Param        body  body      pve.CreateVmRequest  true  "创建参数（vmid、name、cores、memory、diskSource、storage 为必填；CloudInit 字段均为可选）"
// @Success      202   {object}  map[string]any
// @Failure      400   {object}  map[string]string
// @Failure      404   {object}  map[string]string
// @Failure      409   {object}  map[string]string
// @Failure      500   {object}  map[string]string
// @Failure      502   {object}  map[string]string
// @Router       /nodes/{node}/vms [post]
func createVm(c *gin.Context) {
	node := c.Param("node")
	var req pve.CreateVmRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}
	if req.VMID == 0 || req.Name == "" || req.Cores == 0 || req.Memory == 0 || req.DiskSource == "" || req.Storage == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vmid, name, cores, memory, diskSource and storage are required"})
		return
	}
	data, err := vmsSvcGlobal.CreateVm(node, req)
	if err != nil {
		handlePveError(c, err)
		return
	}
	var result any
	json.Unmarshal(data, &result)
	c.Header("Location", "/api/pve/nodes/"+node+"/vms/"+fmt.Sprint(req.VMID))
	respondAccepted(c, result)
}

// listStorage godoc
// @Summary      列出所有存储
// @Description  返回 PVE 中所有存储资源列表
// @Tags         storage
// @Produce      json
// @Success      200  {object}  map[string]any
// @Failure      500  {object}  map[string]string
// @Failure      502  {object}  map[string]string
// @Router       /storage [get]
func listStorage(c *gin.Context) {
	storages, err := storageSvcGlobal.ListStorage()
	if err != nil {
		handlePveError(c, err)
		return
	}
	respondOK(c, storages)
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
