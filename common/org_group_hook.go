package common

// OrgMemberGroupGuard is an optional hook set by Lynxton to enforce that an
// enterprise sub-account (org member) may only relay through channel groups that
// the organization has assigned to it (lc_org_members.usable_groups).
//
// It is called in the relay distributor after the using-group is finalized
// (covering both the token group and any playground override), receiving the
// authenticated user id and the finalized using-group. Returning a non-nil error
// denies the request; its message is surfaced to the client. A nil hook (the
// default) means no enforcement. Mirrors the EmitAsyncBillingOpsLog pattern.
var OrgMemberGroupGuard func(userID int, usingGroup string) error
