package common

import "github.com/gin-gonic/gin"

// ModelListHook lets Lynxton replace the /v1/models listing with a catalog-aware implementation
// without importing internal packages from upstream controller code.
var ModelListHook func(c *gin.Context, modelType int)
