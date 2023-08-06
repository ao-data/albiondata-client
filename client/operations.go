package client

type operation interface {
	Process(state *albionState)
}

// Notes:
//   2020-08-31 (@phendryx): opAuctionGetItemsAverage removed from op codes
//			     based on public suggested changes and
//               @marleythemongolianmoose's findings:
//               "MarleyTheMongolianMoose: AuctionGetItemsAverage == 92 == kind
//               of looks like it disappears in the new one"

// OperationType used to identify operation types
//
//go:generate stringer -type=OperationType
type OperationType uint16

const (
	opUnused OperationType = iota
	opPing
	opJoin
	opVersionedOperation
	opCreateAccount
	opLogin
	opCreateGuestAccount
	opSendCrashLog
	opSendTraceRoute
	opSendVfxStats
	opSendGamePingInfo
	opCreateCharacter
	opDeleteCharacter
	opSelectCharacter
	opAcceptPopups
	opRedeemKeycode
	opGetGameServerByCluster
	opGetShopPurchaseUrl
	opGetReferralSeasonDetails
	opGetReferralLink
	opGetShopTilesForCategory
	opMove
	opAttackStart
	opCastStart
	opCastCancel
	opTerminateToggleSpell
	opChannelingCancel
	opAttackBuildingStart
	opInventoryDestroyItem
	opInventoryMoveItem
	opInventoryRecoverItem
	opInventoryRecoverAllItems
	opInventorySplitStack
	opInventorySplitStackInto
	opGetClusterData
	opChangeCluster
	opConsoleCommand
	opChatMessage
	opReportClientError
	opRegisterToObject
	opUnRegisterFromObject
	opCraftBuildingChangeSettings
	opCraftBuildingTakeMoney
	opRepairBuildingChangeSettings
	opRepairBuildingTakeMoney
	opActionBuildingChangeSettings
	opHarvestStart
	opHarvestCancel
	opTakeSilver
	opActionOnBuildingStart
	opActionOnBuildingCancel
	opInstallResourceStart
	opInstallResourceCancel
	opInstallSilver
	opBuildingFillNutrition
	opBuildingChangeRenovationState
	opBuildingBuySkin
	opBuildingClaim
	opBuildingGiveup
	opBuildingNutritionSilverStorageDeposit
	opBuildingNutritionSilverStorageWithdraw
	opBuildingNutritionSilverRewardSet
	opConstructionSiteCreate
	opPlaceableObjectPlace
	opPlaceableObjectPlaceCancel
	opPlaceableObjectPickup
	opFurnitureObjectUse
	opFarmableHarvest
	opFarmableFinishGrownItem
	opFarmableDestroy
	opFarmableGetProduct
	opFarmableFill
	opTearDownConstructionSite
	opCastleGateUse
	opAuctionCreateOffer
	opAuctionCreateRequest
	opAuctionGetOffers
	opAuctionGetRequests
	opAuctionBuyOffer
	opAuctionAbortAuction
	opAuctionModifyAuction
	opAuctionAbortOffer
	opAuctionAbortRequest
	opAuctionSellRequest
	opAuctionGetFinishedAuctions
	opAuctionGetFinishedAuctionsCount
	opAuctionFetchAuction
	opAuctionGetMyOpenOffers
	opAuctionGetMyOpenRequests
	opAuctionGetMyOpenAuctions
	opAuctionGetItemAverageStats
	opAuctionGetItemAverageValue
	opContainerOpen
	opContainerClose
	opContainerManageSubContainer
	opRespawn
	opSuicide
	opJoinGuild
	opLeaveGuild
	opCreateGuild
	opInviteToGuild
	opDeclineGuildInvitation
	opKickFromGuild
	opInstantJoinGuild
	opDuellingChallengePlayer
	opDuellingAcceptChallenge
	opDuellingDenyChallenge
	opChangeClusterTax
	opClaimTerritory
	opGiveUpTerritory
	opChangeTerritoryAccessRights
	opGetMonolithInfo
	opGetClaimInfo
	opGetAttackInfo
	opGetTerritorySeasonPoints
	opGetAttackSchedule
	opScheduleAttack
	opGetMatches
	opGetMatchDetails
	opJoinMatch
	opLeaveMatch
	opChangeChatSettings
	opLogoutStart
	opLogoutCancel
	opClaimOrbStart
	opClaimOrbCancel
	opMatchLootChestOpeningStart
	opMatchLootChestOpeningCancel
	opDepositToGuildAccount
	opWithdrawalFromAccount
	opChangeGuildPayUpkeepFlag
	opChangeGuildTax
	opGetMyTerritories
	opMorganaCommand
	opGetServerInfo
	opSubscribeToCluster
	opAnswerMercenaryInvitation
	opGetCharacterEquipment
	opGetCharacterSteamAchievements
	opGetCharacterStats
	opGetKillHistoryDetails
	opLearnMasteryLevel
	opReSpecAchievement
	opChangeAvatar
	opGetRankings
	opGetRank
	opGetGvgSeasonRankings
	opGetGvgSeasonRank
	opGetGvgSeasonHistoryRankings
	opGetGvgSeasonGuildMemberHistory
	opKickFromGvGMatch
	opGetCrystalLeagueDailySeasonPoints
	opGetChestLogs
	opGetAccessRightLogs
	opGetGuildAccountLogs
	opGetGuildAccountLogsLargeAmount
	opInviteToPlayerTrade
	opPlayerTradeCancel
	opPlayerTradeInvitationAccept
	opPlayerTradeAddItem
	opPlayerTradeRemoveItem
	opPlayerTradeAcceptTrade
	opPlayerTradeSetSilverOrGold
	opSendMiniMapPing
	opStuck
	opBuyRealEstate
	opClaimRealEstate
	opGiveUpRealEstate
	opChangeRealEstateOutline
	opGetMailInfos
	opGetMailCount
	opReadMail
	opSendNewMail
	opDeleteMail
	opMarkMailUnread
	opClaimAttachmentFromMail
	opApplyToGuild
	opAnswerGuildApplication
	opRequestGuildFinderFilteredList
	opUpdateGuildRecruitmentInfo
	opRequestGuildRecruitmentInfo
	opRequestGuildFinderNameSearch
	opRequestGuildFinderRecommendedList
	opRegisterChatPeer
	opSendChatMessage
	opSendModeratorMessage
	opJoinChatChannel
	opLeaveChatChannel
	opSendWhisperMessage
	opSay
	opPlayEmote
	opStopEmote
	opGetClusterMapInfo
	opAccessRightsChangeSettings
	opMount
	opMountCancel
	opBuyJourney
	opSetSaleStatusForEstate
	opResolveGuildOrPlayerName
	opGetRespawnInfos
	opMakeHome
	opLeaveHome
	opResurrectionReply
	opAllianceCreate
	opAllianceDisband
	opAllianceGetMemberInfos
	opAllianceInvite
	opAllianceAnswerInvitation
	opAllianceCancelInvitation
	opAllianceKickGuild
	opAllianceLeave
	opAllianceChangeGoldPaymentFlag
	opAllianceGetDetailInfo
	opGetIslandInfos
	opAbandonMyIsland
	opBuyMyIsland
	opBuyGuildIsland
	opAbandonGuildIsland
	opUpgradeMyIsland
	opUpgradeGuildIsland
	opMoveMyIsland
	opMoveGuildIsland
	opTerritoryFillNutrition
	opTeleportBack
	opPartyInvitePlayer
	opPartyRequestJoin
	opPartyAnswerInvitation
	opPartyAnswerJoinRequest
	opPartyLeave
	opPartyKickPlayer
	opPartyMakeLeader
	opPartyChangeLootSetting
	opPartyMarkObject
	opPartySetRole
	opSetGuildCodex
	opExitEnterStart
	opExitEnterCancel
	opQuestGiverRequest
	opGoldMarketGetBuyOffer
	opGoldMarketGetBuyOfferFromSilver
	opGoldMarketGetSellOffer
	opGoldMarketGetSellOfferFromSilver
	opGoldMarketBuyGold
	opGoldMarketSellGold
	opGoldMarketCreateSellOrder
	opGoldMarketCreateBuyOrder
	opGoldMarketGetInfos
	opGoldMarketCancelOrder
	opGoldMarketGetAverageInfo
	opTreasureChestUsingStart
	opTreasureChestUsingCancel
	opUseLootChest
	opUseShrine
	opUseHellgateShrine
	opLaborerStartJob
	opLaborerTakeJobLoot
	opLaborerDismiss
	opLaborerMove
	opLaborerBuyItem
	opLaborerUpgrade
	opBuyPremium
	opRealEstateGetAuctionData
	opRealEstateBidOnAuction
	opFriendInvite
	opFriendAnswerInvitation
	opFriendCancelnvitation
	opFriendRemove
	opInventoryStack
	opInventorySort
	opInventoryDropAll
	opInventoryAddToStacks
	opEquipmentItemChangeSpell
	opExpeditionRegister
	opExpeditionRegisterCancel
	opJoinExpedition
	opDeclineExpeditionInvitation
	opVoteStart
	opVoteDoVote
	opRatingDoRate
	opEnteringExpeditionStart
	opEnteringExpeditionCancel
	opActivateExpeditionCheckPoint
	opArenaRegister
	opArenaAddInvite
	opArenaRegisterCancel
	opArenaLeave
	opJoinArenaMatch
	opDeclineArenaInvitation
	opEnteringArenaStart
	opEnteringArenaCancel
	opArenaCustomMatch
	opUpdateCharacterStatement
	opBoostFarmable
	opGetStrikeHistory
	opUseFunction
	opUsePortalEntrance
	opResetPortalBinding
	opQueryPortalBinding
	opClaimPaymentTransaction
	opChangeUseFlag
	opClientPerformanceStats
	opExtendedHardwareStats
	opClientLowMemoryWarning
	opTerritoryClaimStart
	opTerritoryClaimCancel
	opClaimPowerCrystalStart
	opClaimPowerCrystalCancel
	opTerritoryUpgradeWithPowerCrystal
	opRequestAppStoreProducts
	opVerifyProductPurchase
	opQueryGuildPlayerStats
	opQueryAllianceGuildStats
	opTrackAchievements
	opSetAchievementsAutoLearn
	opDepositItemToGuildCurrency
	opWithdrawalItemFromGuildCurrency
	opAuctionSellSpecificItemRequest
	opFishingStart
	opFishingCasting
	opFishingCast
	opFishingCatch
	opFishingPull
	opFishingGiveLine
	opFishingFinish
	opFishingCancel
	opCreateGuildAccessTag
	opDeleteGuildAccessTag
	opRenameGuildAccessTag
	opFlagGuildAccessTagGuildPermission
	opAssignGuildAccessTag
	opRemoveGuildAccessTagFromPlayer
	opModifyGuildAccessTagEditors
	opRequestPublicAccessTags
	opChangeAccessTagPublicFlag
	opUpdateGuildAccessTag
	opSteamStartMicrotransaction
	opSteamFinishMicrotransaction
	opSteamIdHasActiveAccount
	opCheckEmailAccountState
	opLinkAccountToSteamId
	opInAppConfirmPaymentGooglePlay
	opInAppConfirmPaymentAppleAppStore
	opInAppPurchaseRequest
	opInAppPurchaseFailed
	opCharacterSubscriptionInfo
	opAccountSubscriptionInfo
	opBuyGvgSeasonBooster
	opChangeFlaggingPrepare
	opOverCharge
	opOverChargeEnd
	opRequestTrusted
	opChangeGuildLogo
	opPartyFinderRegisterForUpdates
	opPartyFinderUnregisterForUpdates
	opPartyFinderEnlistNewPartySearch
	opPartyFinderDeletePartySearch
	opPartyFinderChangePartySearch
	opPartyFinderChangeRole
	opPartyFinderApplyForGroup
	opPartyFinderAcceptOrDeclineApplyForGroup
	opPartyFinderGetEquipmentSnapshot
	opPartyFinderRegisterApplicants
	opPartyFinderUnregisterApplicants
	opPartyFinderFulltextSearch
	opPartyFinderRequestEquipmentSnapshot
	opGetPersonalSeasonTrackerData
	opGetPersonalSeasonPastRewardData
	opUseConsumableFromInventory
	opClaimPersonalSeasonReward
	opEasyAntiCheatMessageToServer
	opXignCodeMessageToServer
	opBattlEyeMessageToServer
	opSetNextTutorialState
	opAddPlayerToMuteList
	opRemovePlayerFromMuteList
	opProductShopUserEvent
	opGetVanityUnlocks
	opBuyVanityUnlocks
	opGetMountSkins
	opSetMountSkin
	opSetWardrobe
	opChangeCustomization
	opChangePlayerIslandData
	opGetGuildChallengePoints
	opSmartQueueJoin
	opSmartQueueLeave
	opSmartQueueSelectSpawnCluster
	opUpgradeHideout
	opInitHideoutAttackStart
	opInitHideoutAttackCancel
	opHideoutFillNutrition
	opHideoutGetInfo
	opHideoutGetOwnerInfo
	opHideoutSetTribute
	opHideoutUpgradeWithPowerCrystal
	opHideoutDeclareHQ
	opHideoutUndeclareHQ
	opHideoutGetHQRequirements
	opHideoutBoost
	opHideoutBoostConstruction
	opOpenWorldAttackScheduleStart
	opOpenWorldAttackScheduleCancel
	opOpenWorldAttackConquerStart
	opOpenWorldAttackConquerCancel
	opGetOpenWorldAttackDetails
	opGetNextOpenWorldAttackScheduleTime
	opRecoverVaultFromHideout
	opGetGuildEnergyDrainInfo
	opChannelingUpdate
	opUseCorruptedShrine
	opRequestEstimatedMarketValue
	opLogFeedback
	opGetInfamyInfo
	opGetPartySmartClusterQueuePriority
	opSetPartySmartClusterQueuePriority
	opClientAntiAutoClickerInfo
	opClientBotPatternDetectionInfo
	opClientAntiGatherClickerInfo
	opLoadoutCreate
	opLoadoutRead
	opLoadoutReadHeaders
	opLoadoutUpdate
	opLoadoutDelete
	opLoadoutOrderUpdate
	opLoadoutEquip
	opBatchUseItemCancel
	opEnlistFactionWarfare
	opGetFactionWarfareWeeklyReport
	opClaimFactionWarfareWeeklyReport
	opGetFactionWarfareCampaignData
	opClaimFactionWarfareItemReward
	opSendMemoryConsumption
	opPickupPowerCrystalStart
	opPickupPowerCrystalCancel
	opSetSavingChestLogsFlag
	opGetSavingChestLogsFlag
	opRegisterGuestAccount
	opResendGuestAccountVerificationEmail
	opDoSimpleActionStart
	opDoSimpleActionCancel
	opGetGvgSeasonContributionByActivity
	opGetGvgSeasonContributionByCrystalLeague
	opGetGuildMightCategoryContribution
	opGetGuildMightCategoryOverview
	opGetPvpChallengeData
	opClaimPvpChallengeWeeklyReward
	opGetPersonalMightStats
	opGetGvgSeasonGuildParticipationTime
	opAuctionGetLoadoutOffers
	opAuctionBuyLoadoutOffer
	opAccountDeletionRequest
	opAccountReactivationRequest
	opGetModerationEscalationDefiniton
	opEventBasedPopupAddSeen
	opGetItemKillHistory
	opGetVanityConsumables
	opEquipKillEmote
	opChangeKillEmotePlayOnKnockdownSetting
	opBuyVanityConsumableCharges
	opGetArenaRankings
	opGetCrystalLeagueStatistics
	opSendOptionsLog
	opSendControlsOptionsLog
	opMistsUseImmediateReturnExit
	opMistsUseStaticEntrance
	opMistsUseCityRoadsEntrance
	opChangeNewGuildMemberMail
	opGetNewGuildMemberMail
	opChangeGuildFactionAllegiance
	opGetGuildFactionAllegiance
	opGuildBannerChange
	opGuildGetOptionalStats
	opGuildSetOptionalStats
	opGetPlayerInfoForStalk
	opPayGoldForCharacterTypeChange
	opQuickSellAuctionQueryAction
	opQuickSellAuctionSellAction
	opFcmTokenToServer
	opApnsTokenToServer
	opDeathRecap
	opAuctionFetchFinishedAuctions
	opAbortAuctionFetchFinishedAuctions
	opRequestLegendaryEvenHistory
	opPartyAnswerStartHuntRequest
	opHuntAbort
)
