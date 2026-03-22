package main

import (
	"encoding/json"
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
// @Tags         nodes
// @Produce      json
// @Param        node  path      string  true  "节点名称"
// @Success      200   {object}  map[string]any
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
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
// @Tags         vms
// @Produce      json
// @Param        node  path      string  true  "节点名称"
// @Success      200   {object}  map[string]any
// @Failure      500   {object}  map[string]string
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
// @Tags         vms
// @Produce      json
// @Param        node  path      string  true  "节点名称"
// @Param        vmid  path      string  true  "虚拟机 ID"
// @Success      200   {object}  map[string]any
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
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
// @Tags         vms
// @Produce      json
// @Param        node  path      string  true  "节点名称"
// @Param        vmid  path      string  true  "虚拟机 ID"
// @Success      202   {object}  map[string]any
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
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
// @Tags         vms
// @Produce      json
// @Param        node  path      string  true  "节点名称"
// @Param        vmid  path      string  true  "虚拟机 ID"
// @Success      202   {object}  map[string]any
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
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
// @Tags         vms
// @Produce      json
// @Param        node  path      string  true  "节点名称"
// @Param        vmid  path      string  true  "虚拟机 ID"
// @Success      202   {object}  map[string]any
// @Failure      404   {object}  map[string]string
// @Failure      500   {object}  map[string]string
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

// listStorage godoc
// @Summary      列出所有存储
// @Tags         storage
// @Produce      json
// @Success      200  {object}  map[string]any
// @Failure      500  {object}  map[string]string
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
