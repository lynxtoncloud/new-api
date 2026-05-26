package constant

type ContextKey string

const (
	ContextKeyTokenCountMeta  ContextKey = "token_count_meta"
	ContextKeyPromptTokens    ContextKey = "prompt_tokens"
	ContextKeyEstimatedTokens ContextKey = "estimated_tokens"

	ContextKeyOriginalModel    ContextKey = "original_model"
	ContextKeyRequestStartTime ContextKey = "request_start_time"

	/* token related keys */
	ContextKeyTokenUnlimited         ContextKey = "token_unlimited_quota"
	ContextKeyTokenKey               ContextKey = "token_key"
	ContextKeyTokenId                ContextKey = "token_id"
	ContextKeyTokenGroup             ContextKey = "token_group"
	ContextKeyTokenSpecificChannelId ContextKey = "specific_channel_id"
	ContextKeyTokenModelLimitEnabled ContextKey = "token_model_limit_enabled"
	ContextKeyTokenModelLimit        ContextKey = "token_model_limit"
	ContextKeyTokenCrossGroupRetry   ContextKey = "token_cross_group_retry"

	/* channel related keys */
	ContextKeyChannelId                ContextKey = "channel_id"
	ContextKeyChannelName              ContextKey = "channel_name"
	ContextKeyChannelCreateTime        ContextKey = "channel_create_time"
	ContextKeyChannelBaseUrl           ContextKey = "base_url"
	ContextKeyChannelType              ContextKey = "channel_type"
	ContextKeyChannelSetting           ContextKey = "channel_setting"
	ContextKeyChannelOtherSetting      ContextKey = "channel_other_setting"
	ContextKeyChannelParamOverride     ContextKey = "param_override"
	ContextKeyChannelHeaderOverride    ContextKey = "header_override"
	ContextKeyChannelOrganization      ContextKey = "channel_organization"
	ContextKeyChannelAutoBan           ContextKey = "auto_ban"
	ContextKeyChannelModelMapping      ContextKey = "model_mapping"
	ContextKeyChannelStatusCodeMapping ContextKey = "status_code_mapping"
	ContextKeyChannelIsMultiKey        ContextKey = "channel_is_multi_key"
	ContextKeyChannelMultiKeyIndex     ContextKey = "channel_multi_key_index"
	ContextKeyChannelKey               ContextKey = "channel_key"

	ContextKeyAutoGroup           ContextKey = "auto_group"
	ContextKeyAutoGroupIndex      ContextKey = "auto_group_index"
	ContextKeyAutoGroupRetryIndex ContextKey = "auto_group_retry_index"

	/* alias routing keys (Lynxton model-catalog-design.md §7/§13)
	 *
	 * 解析阶段（distributor）若识别为 alias，则注入以下 key；下游 RelayInfo / model_mapped / relay 重试
	 * 按 key 存在与否区分 alias / direct 两条路径。命名与 override 的 "upstream_model" 显式区分，
	 * 不复用 ContextKeyOriginalModel 以避免与现有计费 / 重试契约耦合。
	 */
	ContextKeyAliasOriginModel    ContextKey = "alias_origin_model"     // = 对外 alias model_id，全程不变
	ContextKeyAliasUpstreamModel  ContextKey = "alias_upstream_model"   // = 当前选中的真实 target_model_id
	ContextKeyAliasOrderedTargets ContextKey = "alias_ordered_targets"  // []string，第一段排好序的候选游标
	ContextKeyAliasTargetCursor   ContextKey = "alias_target_cursor"    // int，当前 target 在 ordered list 中的索引

	/* user related keys */
	ContextKeyUserId      ContextKey = "id"
	ContextKeyUserSetting ContextKey = "user_setting"
	ContextKeyUserQuota   ContextKey = "user_quota"
	ContextKeyUserStatus  ContextKey = "user_status"
	ContextKeyUserEmail   ContextKey = "user_email"
	ContextKeyUserGroup   ContextKey = "user_group"
	ContextKeyUsingGroup  ContextKey = "group"
	ContextKeyUserName    ContextKey = "username"

	ContextKeyLocalCountTokens ContextKey = "local_count_tokens"

	ContextKeySystemPromptOverride ContextKey = "system_prompt_override"

	// ContextKeyFileSourcesToCleanup stores file sources that need cleanup when request ends
	ContextKeyFileSourcesToCleanup ContextKey = "file_sources_to_cleanup"

	// ContextKeyAdminRejectReason stores an admin-only reject/block reason extracted from upstream responses.
	// It is not returned to end users, but can be persisted into consume/error logs for debugging.
	ContextKeyAdminRejectReason ContextKey = "admin_reject_reason"

	// ContextKeyLanguage stores the user's language preference for i18n
	ContextKeyLanguage ContextKey = "language"
	ContextKeyIsStream ContextKey = "is_stream"

	/* enterprise discount related keys */
	ContextKeyOrgID           ContextKey = "org_id"
	ContextKeyOrgDiscountRate ContextKey = "org_discount_rate"

	// Ops billing snapshot (set by new-api/service before RecordConsumeLog; read by Lynxton relay_request_summary).
	ContextKeyOpsBillingPromptTokens     ContextKey = "ops_billing_prompt_tokens"
	ContextKeyOpsBillingCompletionTokens ContextKey = "ops_billing_completion_tokens"
	ContextKeyOpsBillingTotalTokens      ContextKey = "ops_billing_total_tokens"
	ContextKeyOpsBillingConsumeQuota     ContextKey = "ops_billing_consume_quota"
	ContextKeyOpsBillingSnapshotFrom     ContextKey = "ops_billing_snapshot_from"

	// HappyHorse relay diagnostics (set in ali adaptor BuildRequestBody; read by Lynxton relay_request_summary).
	ContextKeyOpsHappyHorseVariant          ContextKey = "ops_happyhorse_variant"
	ContextKeyOpsHappyHorseClientModel      ContextKey = "ops_happyhorse_client_model"
	ContextKeyOpsHappyHorseUpstreamModel    ContextKey = "ops_happyhorse_upstream_model"
	ContextKeyOpsHappyHorseImagesCount      ContextKey = "ops_happyhorse_images_count"
	ContextKeyOpsHappyHorseImagesBytes      ContextKey = "ops_happyhorse_images_bytes"
	ContextKeyOpsHappyHorseMediaCount       ContextKey = "ops_happyhorse_media_count"
	ContextKeyOpsHappyHorseMediaTypes       ContextKey = "ops_happyhorse_media_types"
	ContextKeyOpsHappyHorseUpstreamBodyBytes ContextKey = "ops_happyhorse_upstream_body_bytes"
	ContextKeyOpsHappyHorseFinalizeError    ContextKey = "ops_happyhorse_finalize_error"
	ContextKeyOpsHappyHorseSnapshotFrom     ContextKey = "ops_happyhorse_snapshot_from"
)
