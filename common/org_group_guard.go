package common

// OrgMemberGroupGuard 企业成员可用分组兜底（由 Lynxton 注入；nil 表示不校验）。
var OrgMemberGroupGuard func(userID int, usingGroup string) error
