package client

// returnCodeNames maps Photon OperationResponseCode (cast to int16) to a human-readable name.
// Values sourced from Albion.Common.Photon Rc* enums in the IL2CPP dump.
var returnCodeNames = map[int16]string{
	// RcGeneric
	0:    "OK",
	-1:   "UnhandledOperation",
	-500: "InternalServerError",
	-501: "InvalidState",
	-502: "CharacterIsDead",
	-503: "InvalidTimeStamp",
	-504: "IncorrectClientVersion",

	// RcOpLogin
	200: "Login_IncorrectClientVersion",
	201: "Login_InvalidAccountNameOrPassword",
	202: "Login_InvalidRequestAlreadyLoggedIn",
	203: "Login_AccountAlreadyInUse",
	204: "Login_InsufficientAccessLevel",
	205: "Login_AccountIsBanned",
	206: "Login_LoginServiceUnavailable",
	207: "Login_InvalidEmailAddressFormat",
	208: "Login_ClientCorrupted",
	209: "Login_AccountControlledByCommunityManagement",
	210: "Login_BlockedByEnteringWrongPassword",
	211: "Login_OriginVerificationNeeded",
	212: "Login_OriginVerificationFailed",
	213: "Login_WaitingForEmailVerification",
	214: "Login_SteamIdGivenIsDifferentFromLinkedSteamId",
	215: "Login_EpicIdGivenIsDifferentFromLinkedEpicId",
	216: "Login_XboxIdGivenIsDifferentFromLinkedXboxId",
	217: "Login_LoggedInSteamIdNotLinkedToAnAlbionAccount",
	218: "Login_SteamDidNotVerifySessionTicket",
	219: "Login_EpicDidNotVerifyAccessToken",
	220: "Login_XSTSTokenInvalid",
	221: "Login_XSTSTokenNotYetValid",
	222: "Login_XSTSTokenExpired",
	223: "Login_InvalidGuestAccount",
	224: "Login_InvalidGuestAccountDevice",
	225: "Login_LoginNotPossibleNotAGuestAccount",
	226: "Login_LoginNotPossibleAffiliateAccount",
	227: "Login_WrongXmlContext",
	228: "Login_FounderStateTooLowForEarlyAccess",

	// RcOpChangeCluster
	1640: "ChangeCluster_NotNearAnExit",
	1641: "ChangeCluster_ExitNotConnected",
	1642: "ChangeCluster_ExitInfoNotFound",
	1643: "ChangeCluster_NotInGuildOfThisTerritory",
	1644: "ChangeCluster_NotInGuildOfTargetTerritory",
	1645: "ChangeCluster_ThisTerritoryWithoutOwner",
	1646: "ChangeCluster_TargetTerritoryWithoutOwner",
	1647: "ChangeCluster_TooEarlyWaitForExitCooldown",
	1648: "ChangeCluster_CannotUseExitWhileInCombat",
	1649: "ChangeCluster_CannotUseExitWhileInClusterQueue",
	1650: "ChangeCluster_TerritoryDoesNotExist",
	1651: "ChangeCluster_ClusterChangeAlreadyInProgress",
	1652: "ChangeCluster_ReputationTooLow",
	1653: "ChangeCluster_KnockedDownOrDead",
	1654: "ChangeCluster_NoOutlawsAllowed",
	1655: "ChangeCluster_NoFactionPlayersAllowed",
	1656: "ChangeCluster_NoOpposingFactionPlayersAllowed",
	1657: "ChangeCluster_ClusterIsFull",
	1658: "ChangeCluster_ClusterIsFullWithQueue",
	1659: "ChangeCluster_UnderCannotUseExitSpell",
	1660: "ChangeCluster_NoChargesLeft",
	1661: "ChangeCluster_OnCooldownAfterSmartQueueJoin",
	1662: "ChangeCluster_AlternativeClusterUnavailable",
	1663: "ChangeCluster_SkipTooEarlyPersonalCooldown",
	1664: "ChangeCluster_TooEarlyQueueNotActiveLongEnough",
	1665: "ChangeCluster_AvalonRoadsNoFactionEnterAllowed",
	1666: "ChangeCluster_MistCityRequirementsNotMet",
	1667: "ChangeCluster_DisabledByFeatureSwitch",
	1668: "ChangeCluster_CouldNotEnterExitOtherActionStillActive",
	1669: "ChangeCluster_AlreadyEnteringExit",
	1670: "ChangeCluster_ClusterIsAlreadyShuttingDown",
	1671: "ChangeCluster_UsageBlockedByObjectBlockerBuff",
	1672: "ChangeCluster_SilverCostsChanged",
	1673: "ChangeCluster_NoEnoughSilver",
	1674: "ChangeCluster_NoOutlandsClusterFound",
	1675: "ChangeCluster_NoExitFoundInOutlandCluster",
	1676: "ChangeCluster_ExitLocked",

	// RcOpChangeChatSettings
	5080: "ChangeChatSettings_InvalidAction",
	5081: "ChangeChatSettings_InvalidPlayer",
	5082: "ChangeChatSettings_InvalidTab",
	5083: "ChangeChatSettings_TabExists",
	5084: "ChangeChatSettings_TabDontExist",
	5085: "ChangeChatSettings_MaxTabs",
	5086: "ChangeChatSettings_InvalidChannels",

	// RcOpGetSiegeBannerInfo
	10240: "GetSiegeBannerInfo_ItemNotFound",
	10241: "GetSiegeBannerInfo_NoGuild",
	10242: "GetSiegeBannerInfo_InvalidItem",
	10243: "GetSiegeBannerInfo_NoMonolithFound",
	10244: "GetSiegeBannerInfo_OwnGuild",
	10245: "GetSiegeBannerInfo_NoOwner",

	// RcOpFriendInvite
	10640: "FriendInvite_PlayerNotFound",
	10641: "FriendInvite_CantInviteYourself",
	10642: "FriendInvite_TargetPlayerIsAlreadyInvited",
	10643: "FriendInvite_TargetPlayerIsInTutorial",
	10644: "FriendInvite_FriendLimitReached",
	10645: "FriendInvite_FriendLimitReachedForOtherPlayer",
	10646: "FriendInvite_FriendRequestLimitReached",
	10647: "FriendInvite_FriendRequestLimitReachedForOtherPlayer",
	10648: "FriendInvite_PlayerIsAlreadyFriend",
	10649: "FriendInvite_PlayerWasAddedAsFriend",

	// RcOpPartyFinderEnlistNewPartySearch
	14320: "PartyFinderEnlist_InfoTooLong",
	14321: "PartyFinderEnlist_TitleTooLong",
	14322: "PartyFinderEnlist_InvalidGroupSize",
	14323: "PartyFinderEnlist_InvalidItemPower",
	14324: "PartyFinderEnlist_InPartyButNotLeader",
	14325: "PartyFinderEnlist_ErrorOffset",

	// RcOpPartyFinderApplyForGroup
	14480: "PartyFinderApply_LeaderNotKnown",
	14481: "PartyFinderApply_ApplicantNotKnown",
	14482: "PartyFinderApply_GroupNotKnown",
	14483: "PartyFinderApply_AlreadyFull",
	14484: "PartyFinderApply_ApplicantsPartyDoesnotFit",
	14485: "PartyFinderApply_ApplicantInGroupButNotLeader",
	14486: "PartyFinderApply_MaxApplicationsReached",
	14487: "PartyFinderApply_TriedToApplyOwnSearch",
	14488: "PartyFinderApply_EnlistedPartiesNotAllowedToJoinAnother",
	14489: "PartyFinderApply_AlreadyInPendingPlayers",
	14490: "PartyFinderApply_ItemPowerTooLow",
	14491: "PartyFinderApply_EnteredPendingPlayers",
	14492: "PartyFinderApply_PlayerBlackListed",

	// RcOpHideoutSetTribute
	15840: "HideoutSetTribute_NoPermission",

	// RcOpHideoutDeclareHQ
	15920: "HideoutDeclareHQ_NoGuildPermission",
	15921: "HideoutDeclareHQ_InCooldown",
	15922: "HideoutDeclareHQ_NotEnoughSeasonPoints",
	15923: "HideoutDeclareHQ_CloseToAttacktime",
	15924: "HideoutDeclareHQ_NotInGuild",
	15925: "HideoutDeclareHQ_PowerCrystalsNotAllowed",

	// RcOpHideoutBoost
	16040: "HideoutBoost_NoGuildPermission",
	16041: "HideoutBoost_NotEnoughSeasonPoints",
	16042: "HideoutBoost_NotInGuild",
	16043: "HideoutBoost_WrongHideoutState",
	16044: "HideoutBoost_PowerCrystalsNotAllowed",
	16045: "HideoutBoost_BoosterInCooldown",

	// RcOpHideoutBoostConstruction
	16080: "HideoutBoostConstruction_NoGuildPermission",
	16081: "HideoutBoostConstruction_NotEnoughSeasonPoints",
	16082: "HideoutBoostConstruction_InCooldown",
	16083: "HideoutBoostConstruction_NotInGuild",
	16084: "HideoutBoostConstruction_WrongHideoutState",
	16085: "HideoutBoostConstruction_PowerCrystalsNotAllowed",
	16086: "HideoutBoostConstruction_BoosterInCooldown",

	// RcOpClaimPvpChallengeWeeklyReward
	17920: "ClaimPvpChallengeWeeklyReward_NeedsBoughtPremium",
	17921: "ClaimPvpChallengeWeeklyReward_RewardAlreadyClaimed",
	17922: "ClaimPvpChallengeWeeklyReward_NotEnoughMightToClaim",
	17923: "ClaimPvpChallengeWeeklyReward_NotAllowedInTutorial",
	17924: "ClaimPvpChallengeWeeklyReward_NoSpaceInInventory",
}

// returnCodeName returns a human-readable name for a Photon OperationResponseCode.
// The uint16 wire value is reinterpreted as int16 to handle negative RcGeneric codes.
func returnCodeName(code uint16) string {
	signed := int16(code)
	if name, ok := returnCodeNames[signed]; ok {
		return name
	}
	return ""
}
