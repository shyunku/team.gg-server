package service

import (
	"team.gg-server/core"
	"team.gg-server/types"
	"time"
)

var (
	GetSupportedPositions = [...]string{
		types.PositionTop,
		types.PositionJungle,
		types.PositionMid,
		types.PositionAdc,
		types.PositionSupport,
	}
	GetPossibleTeamPositions = func() []CustomGameTeamPositionVO {
		return []CustomGameTeamPositionVO{
			{Team: 1, Position: types.PositionTop},
			{Team: 1, Position: types.PositionJungle},
			{Team: 1, Position: types.PositionMid},
			{Team: 1, Position: types.PositionAdc},
			{Team: 1, Position: types.PositionSupport},
			{Team: 2, Position: types.PositionTop},
			{Team: 2, Position: types.PositionJungle},
			{Team: 2, Position: types.PositionMid},
			{Team: 2, Position: types.PositionAdc},
			{Team: 2, Position: types.PositionSupport},
		}
	}
	GetInitialMatchCount = func() int {
		if core.DebugMode {
			return types.LoadInitialMatchCountDev
		}
		return types.LoadInitialMatchCount
	}
	GetLoadMoreMatchCount = func() int {
		if core.DebugMode {
			return types.LoadMoreMatchCountDev
		}
		return types.LoadMoreMatchCount
	}
	GetDataExplorerLoopPeriod = func() time.Duration {
		if core.DebugMode {
			return types.DataExplorerLoopPeriodDev
		}
		return types.DataExplorerLoopPeriod
	}
)
