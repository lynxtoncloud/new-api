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

// TokenGroupSelectableGuard is an optional hook set by Lynxton to enforce that a
// user may only assign a token to a channel group that is user-selectable for
// them: the group must be enabled, flagged user_selectable, and rank below the
// user's own group in the priority ladder (equivalently, present in the group
// dropdown). It is the create/edit-time counterpart to the request-time check
// that the priority ladder already enforces via GroupSpecialUsableGroup.
//
// It is called in token create/edit with the authenticated user id and the
// requested token group. Returning a non-nil error denies the request; its
// message is surfaced to the client. A nil hook (the default) means no
// enforcement. Distinct from OrgMemberGroupGuard, which restricts enterprise
// sub-accounts to their org-assigned groups; both may apply to the same call.
var TokenGroupSelectableGuard func(userID int, group string) error
