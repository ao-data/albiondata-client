package client

// EventType used to identify event types
//
//go:generate stringer -type=EventType
type EventType uint16

const (
	evUnused EventType = iota
	evLeave
	evJoinFinished
	evMove
	evTeleport
	evChangeEquipment
	evHealthUpdate
	evHealthUpdates
	evEnergyUpdate
	evDamageShieldUpdate
	evCraftingFocusUpdate
	evActiveSpellEffectsUpdate
	evResetCooldowns
	evAttack
	evCastStart
	evChannelingUpdate
	evCastCancel
	evCastTimeUpdate
	evCastFinished
	evCastSpell
	evCastSpells
	evCastHit
	evCastHits
	evStoredTargetsUpdate
	evChannelingEnded
	evAttackBuilding
	evInventoryPutItem
	evInventoryDeleteItem
	evInventoryState
	evNewCharacter
	evNewEquipmentItem
	evNewSiegeBannerItem
	evNewSimpleItem
	evNewFurnitureItem
	evNewKillTrophyItem
	evNewJournalItem
	evNewLaborerItem
	evNewEquipmentItemLegendarySoul
	evNewSimpleHarvestableObject
	evNewSimpleHarvestableObjectList
	evNewHarvestableObject
	evNewTreasureDestinationObject
	evTreasureDestinationObjectStatus
	evCloseTreasureDestinationObject
	evNewSilverObject
	evNewBuilding
	evHarvestableChangeState
	evMobChangeState
	evFactionBuildingInfo
	evCraftBuildingInfo
	evRepairBuildingInfo
	evMeldBuildingInfo
	evConstructionSiteInfo
	evPlayerBuildingInfo
	evFarmBuildingInfo
	evTutorialBuildingInfo
	evLaborerObjectInfo
	evLaborerObjectJobInfo
	evMarketPlaceBuildingInfo
	evHarvestStart
	evHarvestCancel
	evHarvestFinished
	evTakeSilver
	evRemoveSilver
	evActionOnBuildingStart
	evActionOnBuildingCancel
	evActionOnBuildingFinished
	evItemRerollQualityFinished
	evInstallResourceStart
	evInstallResourceCancel
	evInstallResourceFinished
	evCraftItemFinished
	evLogoutCancel
	evChatMessage
	evChatSay
	evChatWhisper
	evChatMuted
	evPlayEmote
	evStopEmote
	evSystemMessage
	evUtilityTextMessage
	evUpdateMoney
	evUpdateFame
	evUpdateLearningPoints
	evUpdateReSpecPoints
	evUpdateCurrency
	evUpdateFactionStanding
	evUpdateStanding
	evRespawn
	evServerDebugLog
	evCharacterEquipmentChanged
	evRegenerationHealthChanged
	evRegenerationEnergyChanged
	evRegenerationMountHealthChanged
	evRegenerationCraftingChanged
	evRegenerationHealthEnergyComboChanged
	evRegenerationPlayerComboChanged
	evDurabilityChanged
	evNewLoot
	evAttachItemContainer
	evDetachItemContainer
	evInvalidateItemContainer
	evLockItemContainer
	evGuildUpdate
	evGuildPlayerUpdated
	evInvitedToGuild
	evGuildMemberWorldUpdate
	evUpdateMatchDetails
	evObjectEvent
	evNewMonolithObject
	evMonolithHasBannersPlacedUpdate
	evNewOrbObject
	evNewCastleObject
	evNewSpellEffectArea
	evUpdateSpellEffectArea
	evNewChainSpell
	evUpdateChainSpell
	evNewTreasureChest
	evStartMatch
	evStartArenaMatchInfos
	evEndArenaMatch
	evMatchUpdate
	evActiveMatchUpdate
	evNewMob
	evDebugMobInfo
	evDebugVariablesInfo
	evDebugReputationInfo
	evDebugDiminishingReturnInfo
	evDebugSmartClusterQueueInfo
	evClaimOrbStart
	evClaimOrbFinished
	evClaimOrbCancel
	evOrbUpdate
	evOrbClaimed
	evOrbReset
	evNewWarCampObject
	evNewMatchLootChestObject
	evNewArenaExit
	evGuildMemberTerritoryUpdate
	evInvitedMercenaryToMatch
	evClusterInfoUpdate
	evForcedMovement
	evForcedMovementCancel
	evCharacterStats
	evCharacterStatsKillHistory
	evCharacterStatsDeathHistory
	evCharacterStatsKnockDownHistory
	evCharacterStatsKnockedDownHistory
	evGuildStats
	evKillHistoryDetails
	evItemKillHistoryDetails
	evFullAchievementInfo
	evFinishedAchievement
	evAchievementProgressInfo
	evFullAchievementProgressInfo
	evFullTrackedAchievementInfo
	evFullAutoLearnAchievementInfo
	evQuestGiverQuestOffered
	evQuestGiverDebugInfo
	evConsoleEvent
	evTimeSync
	evChangeAvatar
	evChangeMountSkin
	evGameEvent
	evKilledPlayer
	evDied
	evKnockedDown
	evUnconcious
	evMatchPlayerJoinedEvent
	evMatchPlayerStatsEvent
	evMatchPlayerStatsCompleteEvent
	evMatchTimeLineEventEvent
	evMatchPlayerMainGearStatsEvent
	evMatchPlayerChangedAvatarEvent
	evInvitationPlayerTrade
	evPlayerTradeStart
	evPlayerTradeCancel
	evPlayerTradeUpdate
	evPlayerTradeFinished
	evPlayerTradeAcceptChange
	evMiniMapPing
	evMarketPlaceNotification
	evDuellingChallengePlayer
	evNewDuellingPost
	evDuelStarted
	evDuelEnded
	evDuelDenied
	evDuelRequestCanceled
	evDuelLeftArea
	evDuelReEnteredArea
	evNewRealEstate
	evMiniMapOwnedBuildingsPositions
	evRealEstateListUpdate
	evGuildLogoUpdate
	evGuildLogoChanged
	evPlaceableObjectPlace
	evPlaceableObjectPlaceCancel
	evFurnitureObjectBuffProviderInfo
	evFurnitureObjectCheatProviderInfo
	evFarmableObjectInfo
	evNewUnreadMails
	evMailOperationPossible
	evGuildLogoObjectUpdate
	evStartLogout
	evNewChatChannels
	evJoinedChatChannel
	evLeftChatChannel
	evRemovedChatChannel
	evAccessStatus
	evMounted
	evMountStart
	evMountCancel
	evNewTravelpoint
	evNewIslandAccessPoint
	evNewExit
	evUpdateHome
	evUpdateChatSettings
	evResurrectionOffer
	evResurrectionReply
	evLootEquipmentChanged
	evUpdateUnlockedGuildLogos
	evUpdateUnlockedAvatars
	evUpdateUnlockedAvatarRings
	evUpdateUnlockedBuildings
	evNewIslandManagement
	evNewTeleportStone
	evCloak
	evPartyInvitation
	evPartyJoinRequest
	evPartyJoined
	evPartyDisbanded
	evPartyPlayerJoined
	evPartyChangedOrder
	evPartyPlayerLeft
	evPartyLeaderChanged
	evPartyLootSettingChangedPlayer
	evPartySilverGained
	evPartyPlayerUpdated
	evPartyInvitationAnswer
	evPartyJoinRequestAnswer
	evPartyMarkedObjectsUpdated
	evPartyOnClusterPartyJoined
	evPartySetRoleFlag
	evPartyInviteOrJoinPlayerEquipmentInfo
	evPartyReadyCheckUpdate
	evSpellCooldownUpdate
	evNewHellgateExitPortal
	evNewExpeditionExit
	evNewExpeditionNarrator
	evExitEnterStart
	evExitEnterCancel
	evExitEnterFinished
	evNewQuestGiverObject
	evFullQuestInfo
	evQuestProgressInfo
	evQuestGiverInfoForPlayer
	evFullExpeditionInfo
	evExpeditionQuestProgressInfo
	evInvitedToExpedition
	evExpeditionRegistrationInfo
	evEnteringExpeditionStart
	evEnteringExpeditionCancel
	evRewardGranted
	evArenaRegistrationInfo
	evEnteringArenaStart
	evEnteringArenaCancel
	evEnteringArenaLockStart
	evEnteringArenaLockCancel
	evInvitedToArenaMatch
	evUsingHellgateShrine
	evEnteringHellgateLockStart
	evEnteringHellgateLockCancel
	evPlayerCounts
	evInCombatStateUpdate
	evOtherGrabbedLoot
	evTreasureChestUsingStart
	evTreasureChestUsingFinished
	evTreasureChestUsingCancel
	evTreasureChestUsingOpeningComplete
	evTreasureChestForceCloseInventory
	evLocalTreasuresUpdate
	evLootChestSpawnpointsUpdate
	evPremiumChanged
	evPremiumExtended
	evPremiumLifeTimeRewardGained
	evGoldPurchased
	evLaborerGotUpgraded
	evJournalGotFull
	evJournalFillError
	evFriendRequest
	evFriendRequestInfos
	evFriendInfos
	evFriendRequestAnswered
	evFriendOnlineStatus
	evFriendRequestCanceled
	evFriendRemoved
	evFriendUpdated
	evPartyLootItems
	evPartyLootItemsRemoved
	evReputationUpdate
	evDefenseUnitAttackBegin
	evDefenseUnitAttackEnd
	evDefenseUnitAttackDamage
	evUnrestrictedPvpZoneUpdate
	evUnrestrictedPvpZoneStatus
	evReputationImplicationUpdate
	evNewMountObject
	evMountHealthUpdate
	evMountCooldownUpdate
	evNewExpeditionAgent
	evNewExpeditionCheckPoint
	evExpeditionStartEvent
	evVoteEvent
	evRatingEvent
	evNewArenaAgent
	evBoostFarmable
	evUseFunction
	evNewPortalEntrance
	evNewPortalExit
	evNewRandomDungeonExit
	evWaitingQueueUpdate
	evPlayerMovementRateUpdate
	evObserveStart
	evMinimapZergs
	evMinimapSmartClusterZergs
	evPaymentTransactions
	evPerformanceStatsUpdate
	evOverloadModeUpdate
	evDebugDrawEvent
	evRecordCameraMove
	evRecordStart
	evDeliverCarriableObjectStart
	evDeliverCarriableObjectCancel
	evDeliverCarriableObjectReset
	evDeliverCarriableObjectFinished
	evTerritoryClaimStart
	evTerritoryClaimCancel
	evTerritoryClaimFinished
	evTerritoryScheduleResult
	evTerritoryUpgradeWithPowerCrystalResult
	evReceiveCarriableObjectStart
	evReceiveCarriableObjectFinished
	evUpdateAccountState
	evStartDeterministicRoam
	evGuildFullAccessTagsUpdated
	evGuildAccessTagUpdated
	evGvgSeasonUpdate
	evGvgSeasonCheatCommand
	evSeasonPointsByKillingBooster
	evFishingStart
	evFishingCast
	evFishingCatch
	evFishingFinished
	evFishingCancel
	evNewFloatObject
	evNewFishingZoneObject
	evFishingMiniGame
	evAlbionJournalAchievementCompleted
	evUpdatePuppet
	evChangeFlaggingFinished
	evNewOutpostObject
	evOutpostUpdate
	evOutpostClaimed
	evOverChargeEnd
	evOverChargeStatus
	evPartyFinderFullUpdate
	evPartyFinderUpdate
	evPartyFinderApplicantsUpdate
	evPartyFinderEquipmentSnapshot
	evPartyFinderJoinRequestDeclined
	evNewUnlockedPersonalSeasonRewards
	evPersonalSeasonPointsGained
	evPersonalSeasonPastSeasonDataEvent
	evMatchLootChestOpeningStart
	evMatchLootChestOpeningFinished
	evMatchLootChestOpeningCancel
	evNotifyCrystalMatchReward
	evCrystalRealmFeedback
	evNewLocationMarker
	evNewTutorialBlocker
	evNewTileSwitch
	evNewInformationProvider
	evNewDynamicGuildLogo
	evNewDecoration
	evTutorialUpdate
	evTriggerHintBox
	evRandomDungeonPositionInfo
	evNewLootChest
	evUpdateLootChest
	evLootChestOpened
	evUpdateLootProtectedByMobsWithMinimapDisplay
	evNewShrine
	evUpdateShrine
	evUpdateRoom
	evNewMobSoul
	evNewHellgateShrine
	evUpdateHellgateShrine
	evActivateHellgateExit
	evMutePlayerUpdate
	evShopTileUpdate
	evShopUpdate
	evAntiCheatKick
	evBattlEyeServerMessage
	evUnlockVanityUnlock
	evAvatarUnlocked
	evCustomizationChanged
	evBaseVaultInfo
	evGuildVaultInfo
	evBankVaultInfo
	evRecoveryVaultPlayerInfo
	evRecoveryVaultGuildInfo
	evUpdateWardrobe
	evCastlePhaseChanged
	evGuildAccountLogEvent
	evNewHideoutObject
	evNewHideoutManagement
	evNewHideoutExit
	evInitHideoutAttackStart
	evInitHideoutAttackCancel
	evInitHideoutAttackFinished
	evHideoutManagementUpdate
	evHideoutUpgradeWithPowerCrystalResult
	evIpChanged
	evSmartClusterQueueUpdateInfo
	evSmartClusterQueueActiveInfo
	evSmartClusterQueueKickWarning
	evSmartClusterQueueInvite
	evReceivedGvgSeasonPoints
	evTowerPowerPointUpdate
	evOpenWorldAttackScheduleStart
	evOpenWorldAttackScheduleFinished
	evOpenWorldAttackScheduleCancel
	evOpenWorldAttackConquerStart
	evOpenWorldAttackConquerFinished
	evOpenWorldAttackConquerCancel
	evOpenWorldAttackConquerStatus
	evOpenWorldAttackStart
	evOpenWorldAttackEnd
	evNewRandomResourceBlocker
	evNewHomeObject
	evHideoutObjectUpdate
	evUpdateInfamy
	evMinimapPositionMarkers
	evNewTunnelExit
	evCorruptedDungeonUpdate
	evCorruptedDungeonStatus
	evCorruptedDungeonInfamy
	evHellgateRestrictedAreaUpdate
	evHellgateInfamy
	evHellgateStatus
	evHellgateStatusUpdate
	evHellgateSuspense
	evReplaceSpellSlotWithMultiSpell
	evNewCorruptedShrine
	evUpdateCorruptedShrine
	evCorruptedShrineUsageStart
	evCorruptedShrineUsageCancel
	evExitUsed
	evLinkedToObject
	evLinkToObjectBroken
	evEstimatedMarketValueUpdate
	evStuckCancel
	evDungonEscapeReady
	evFactionWarfareClusterState
	evFactionWarfareHasUnclaimedWeeklyReportsEvent
	evSimpleFeedback
	evSmartClusterQueueSkipClusterError
	evXignCodeEvent
	evBatchUseItemStart
	evBatchUseItemEnd
	evRedZoneEventClusterStatus
	evRedZonePlayerNotification
	evRedZoneWorldEvent
	evFactionWarfareStats
	evUpdateFactionBalanceFactors
	evFactionEnlistmentChanged
	evUpdateFactionRank
	evFactionWarfareCampaignRewardsUnlocked
	evFeaturedFeatureUpdate
	evNewCarriableObject
	evMinimapCrystalPositionMarker
	evCarriedObjectUpdate
	evPickupCarriableObjectStart
	evPickupCarriableObjectCancel
	evPickupCarriableObjectFinished
	evDoSimpleActionStart
	evDoSimpleActionCancel
	evDoSimpleActionFinished
	evNotifyGuestAccountVerified
	evMightAndFavorReceivedEvent
	evWeeklyPvpChallengeRewardStateUpdate
	evNewUnlockedPvpSeasonChallengeRewards
	evStaticDungeonEntrancesDungeonEventStatusUpdates
	evStaticDungeonDungeonValueUpdate
	evStaticDungeonEntranceDungeonEventsAborted
	evInAppPurchaseConfirmedGooglePlay
	evFeatureSwitchInfo
	evPartyJoinRequestAborted
	evPartyInviteAborted
	evPartyStartHuntRequest
	evPartyStartHuntRequested
	evPartyStartHuntRequestAnswer
	evPartyPlayerLeaveScheduled
	evGuildInviteDeclined
	evCancelMultiSpellSlots
	evNewVisualEventObject
	evCastleClaimProgress
	evCastleClaimProgressLogo
	evTownPortalUpdateState
	evTownPortalFailed
	evConsumableVanityChargesAdded
	evFestivitiesUpdate
	evNewBannerObject
	evNewMistsImmediateReturnExit
	evMistsPlayerJoinedInfo
	evNewMistsStaticEntrance
	evNewMistsOpenWorldExit
	evNewTunnelExitTemp
	evNewMistsWispSpawn
	evMistsWispSpawnStateChange
	evNewMistsCityEntrance
	evNewMistsCityRoadsEntrance
	evMistsCityRoadsEntrancePartyStateUpdate
	evMistsCityRoadsEntranceClearStateForParty
	evMistsEntranceDataChanged
	evNewCagedObject
	evCagedObjectStateUpdated
	evEntrancePartyBindingCreated
	evEntrancePartyBindingCleared
	evEntrancePartyBindingInfos
	evNewMistsBorderExit
	evNewMistsDungeonExit
	evLocalQuestInfos
	evLocalQuestStarted
	evLocalQuestActive
	evLocalQuestInactive
	evLocalQuestProgressUpdate
	evNewUnrestrictedPvpZone
	evTemporaryFlaggingStatusUpdate
	evSpellTestPerformanceUpdate
	evTransformation
	evTransformationEnd
	evUpdateTrustlevel
	evRevealHiddenTimeStamps
	evModifyItemTraitFinished
	evRerollItemTraitValueFinished
	evHuntQuestProgressInfo
	evHuntStarted
	evHuntFinished
	evHuntAborted
	evHuntMissionStepStateUpdate
	evNewHuntTrack
	evHuntMissionUpdate
	evHuntQuestMissionProgressUpdate
	evHuntTrackUsed
	evHuntTrackUseableAgain
	evMinimapHuntTrackMarkers
	evNoTracksFound
	evHuntQuestAborted
	evInteractWithTrackStart
	evInteractWithTrackCancel
	evInteractWithTrackFinished
	evNewDynamicCompound
	evLegendaryItemDestroyed
	evAttunementInfo
	evTerritoryClaimRaidedRawEnergyCrystalResult
	evCarriedObjectExpiryWarning
	evCarriedObjectExpired
	evTerritoryRaidStart
	evTerritoryRaidCancel
	evTerritoryRaidFinished
	evTerritoryRaidResult
	evTerritoryMonolithActiveRaidStatus
	evTerritoryMonolithActiveRaidCancelled
	evMonolithEnergyStorageUpdate
	evMonolithNextScheduledOpenWorldAttackUpdate
	evMonolithProtectedBuildingsDamageReductionUpdate
	evNewBuildingBaseEvent
	evNewFortificationBuilding
	evNewCastleGateBuilding
	evBuildingDurabilityUpdate
	evMonolithFortificationPointsUpdate
	evFortificationBuildingUpgradeInfo
	evFortificationBuildingsDamageStateUpdate
	evSiegeNotificationEvent
	evUpdateEnemyWarBannerActive
	evTerritoryAnnouncePlayerEjection
	evCastleGateSwitchUseStarted
	evCastleGateSwitchUseFinished
	evFortificationBuildingWillDowngrade
	evBotCommand
	evJournalAchievementProgressUpdate
	evJournalClaimableRewardUpdate
	evKeySync
	evLocalQuestAreaGone
	evDynamicTemplate
	evDynamicTemplateForcedStateChange
	evNewOutlandsTeleportationPortal
	evNewOutlandsTeleportationReturnPortal
	evOutlandsTeleportationBindingCleared
	evOutlandsTeleportationReturnPortalUpdateEvent
	evPlayerUsedOutlandsTeleportationPortal
	evEncumberedRestricted
	evNewPiledObject
	evPiledObjectStateChanged
	evNewSmugglerCrateDeliveryStation
	evKillRewardedNoFame
	evPickupFromPiledObjectStart
	evPickupFromPiledObjectCancel
	evPickupFromPiledObjectReset
	evPickupFromPiledObjectFinished
	evArmoryActivityChange
	evNewKillTrophyFurnitureBuilding
	evHellDungeonsPlayerJoinedInfo
	evNewTileSwitchTrigger
	evNewMultiRewardObject
	evNewHellDungeonSoulShrineObject
	evHellDungeonSoulShrineStateUpdate
	evNewResurrectionShrine
	evUpdateResurrectionShrine
	evStandTimeFinished
	evEpicAchievementAndStatsUpdate
	evSpectateTargetAfterDeathUpdate
	evSpectateTargetAfterDeathEnded
	evNewHellDungeonUpwardExit
	evNewHellDungeonSoulExit
	evNewHellDungeonDownwardExit
	evNewHellDungeonChestExit
	evNewCorruptedStaticEntrance
	evNewHellDungeonStaticEntrance
	evUpdateHellDungeonStaticEntranceState
	evDebugTriggerHellDungeonShutdownStart
	evFullJournalQuestInfo
	evJournalQuestProgressInfo
	evNewHellDungeonRoomShrineObject
	evHellDungeonRoomShrineStateUpdate
	evSimpleBehaviourBuildingStateUpdate
)
