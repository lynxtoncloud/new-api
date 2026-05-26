package common

// AliasResolver 是 Lynxton 注入的可选钩子：把对外 alias model_id 解析成有序真实 target 候选。
//
// 调用约定（model-catalog-design.md §7 / §13）：
//
//   ok=false           → modelID 不是已启用 alias，distributor 走现有 direct 路径，hook 不影响任何字段。
//   ok=true:
//     rejectReason != "" → modelID 是 alias 但被显式拒绝（如未配 ModelRatio/ModelPrice），
//                          distributor 直接返回 503 + reason，**不进入两段式选 channel**。
//     orderedTargets 非空 → distributor 逐 target 复用 CacheGetRandomSatisfiedChannel 完成两段式。
//     orderedTargets 为空且 rejectReason 为 "" → 视为「alias 无可调度 target」，返回模型不可用。
//
// 镜像 OrgMemberGroupGuard 的 DI 模式：nil 表示无注入，distributor 完全兼容直连路径。
//
// Lynxton 在 cmd/lynxtonapi/main.go 启动时把 internal/modelalias 的 AliasResolveHook 绑到这里。
var AliasResolver func(modelID string) (orderedTargets []string, rejectReason string, ok bool)
