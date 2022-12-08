package helper

type CombatTypeArgument int32

const (
	CombatTypeArgument_COMBAT_NONE                       CombatTypeArgument = 0
	CombatTypeArgument_COMBAT_EVT_BEING_HIT              CombatTypeArgument = 1
	CombatTypeArgument_COMBAT_ANIMATOR_STATE_CHANGED     CombatTypeArgument = 2
	CombatTypeArgument_COMBAT_FACE_TO_DIR                CombatTypeArgument = 3
	CombatTypeArgument_COMBAT_SET_ATTACK_TARGET          CombatTypeArgument = 4
	CombatTypeArgument_COMBAT_RUSH_MOVE                  CombatTypeArgument = 5
	CombatTypeArgument_COMBAT_ANIMATOR_PARAMETER_CHANGED CombatTypeArgument = 6
	CombatTypeArgument_ENTITY_MOVE                       CombatTypeArgument = 7
	CombatTypeArgument_SYNC_ENTITY_POSITION              CombatTypeArgument = 8
	CombatTypeArgument_COMBAT_STEER_MOTION_INFO          CombatTypeArgument = 9
	CombatTypeArgument_COMBAT_FORCE_SET_POS_INFO         CombatTypeArgument = 10
	CombatTypeArgument_COMBAT_COMPENSATE_POS_DIFF        CombatTypeArgument = 11
	CombatTypeArgument_COMBAT_MONSTER_DO_BLINK           CombatTypeArgument = 12
	CombatTypeArgument_COMBAT_FIXED_RUSH_MOVE            CombatTypeArgument = 13
	CombatTypeArgument_COMBAT_SYNC_TRANSFORM             CombatTypeArgument = 14
	CombatTypeArgument_COMBAT_LIGHT_CORE_MOVE            CombatTypeArgument = 15
	CombatTypeArgument_COMBAT_BEING_HEALED_NTF           CombatTypeArgument = 16
	CombatTypeArgument_COMBAT_SKILL_ANCHOR_POSITION_NTF  CombatTypeArgument = 17
	CombatTypeArgument_COMBAT_GRAPPLING_HOOK_MOVE        CombatTypeArgument = 18
)
