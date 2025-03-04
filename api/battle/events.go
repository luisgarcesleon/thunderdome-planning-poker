package battle

import (
	"encoding/json"
	"errors"
)

// UserNudge handles notifying user that they need to vote
func (b *Service) UserNudge(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	msg := createSocketEvent("jab_warrior", EventValue, UserID)

	return msg, nil, false
}

// UserVote handles the participants vote event by setting their vote
// and checks if AutoFinishVoting && AllVoted if so ends voting
func (b *Service) UserVote(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	var msg []byte
	var wv struct {
		VoteValue        string `json:"voteValue"`
		PlanID           string `json:"planId"`
		AutoFinishVoting bool   `json:"autoFinishVoting"`
	}
	json.Unmarshal([]byte(EventValue), &wv)

	Plans, AllVoted := b.db.SetVote(BattleID, UserID, wv.PlanID, wv.VoteValue)

	updatedPlans, _ := json.Marshal(Plans)
	msg = createSocketEvent("vote_activity", string(updatedPlans), UserID)

	if AllVoted && wv.AutoFinishVoting {
		plans, err := b.db.EndPlanVoting(BattleID, wv.PlanID)
		if err != nil {
			return nil, err, false
		}
		updatedPlans, _ := json.Marshal(plans)
		msg = createSocketEvent("voting_ended", string(updatedPlans), "")
	}

	return msg, nil, false
}

// UserVoteRetract handles retracting a user vote
func (b *Service) UserVoteRetract(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	PlanID := EventValue

	plans, err := b.db.RetractVote(BattleID, UserID, PlanID)
	if err != nil {
		return nil, err, false
	}

	updatedPlans, _ := json.Marshal(plans)
	msg := createSocketEvent("vote_retracted", string(updatedPlans), UserID)

	return msg, nil, false
}

// UserPromote handles promoting a user to a leader
func (b *Service) UserPromote(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	leaders, err := b.db.SetBattleLeader(BattleID, EventValue)
	if err != nil {
		return nil, err, false
	}
	leadersJson, _ := json.Marshal(leaders)

	msg := createSocketEvent("leaders_updated", string(leadersJson), "")

	return msg, nil, false
}

// UserDemote handles demoting a user from a leader
func (b *Service) UserDemote(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	leaders, err := b.db.DemoteBattleLeader(BattleID, EventValue)
	if err != nil {
		return nil, err, false
	}
	leadersJson, _ := json.Marshal(leaders)

	msg := createSocketEvent("leaders_updated", string(leadersJson), "")

	return msg, nil, false
}

// UserPromoteSelf handles self-promoting a user to a leader
func (b *Service) UserPromoteSelf(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	leaderCode, err := b.db.GetBattleLeaderCode(BattleID)
	if err != nil {
		return nil, err, false
	}

	if EventValue == leaderCode {
		leaders, err := b.db.SetBattleLeader(BattleID, UserID)
		if err != nil {
			return nil, err, false
		}
		leadersJson, _ := json.Marshal(leaders)

		msg := createSocketEvent("leaders_updated", string(leadersJson), "")

		return msg, nil, false
	} else {
		return nil, errors.New("INCORRECT_LEADER_CODE"), false
	}
}

// UserSpectatorToggle handles toggling user spectator status
func (b *Service) UserSpectatorToggle(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	var st struct {
		Spectator bool `json:"spectator"`
	}
	json.Unmarshal([]byte(EventValue), &st)
	users, err := b.db.ToggleSpectator(BattleID, UserID, st.Spectator)
	if err != nil {
		return nil, err, false
	}
	usersJson, _ := json.Marshal(users)

	msg := createSocketEvent("users_updated", string(usersJson), "")

	return msg, nil, false
}

// PlanVoteEnd handles ending plan voting
func (b *Service) PlanVoteEnd(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	plans, err := b.db.EndPlanVoting(BattleID, EventValue)
	if err != nil {
		return nil, err, false
	}
	updatedPlans, _ := json.Marshal(plans)
	msg := createSocketEvent("voting_ended", string(updatedPlans), "")

	return msg, nil, false
}

// Revise handles editing the battle settings
func (b *Service) Revise(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	var rb struct {
		BattleName           string   `json:"battleName"`
		PointValuesAllowed   []string `json:"pointValuesAllowed"`
		AutoFinishVoting     bool     `json:"autoFinishVoting"`
		PointAverageRounding string   `json:"pointAverageRounding"`
		JoinCode             string   `json:"joinCode"`
		LeaderCode           string   `json:"leaderCode"`
	}
	json.Unmarshal([]byte(EventValue), &rb)

	err := b.db.ReviseBattle(
		BattleID,
		rb.BattleName,
		rb.PointValuesAllowed,
		rb.AutoFinishVoting,
		rb.PointAverageRounding,
		rb.JoinCode,
		rb.LeaderCode,
	)
	if err != nil {
		return nil, err, false
	}

	rb.LeaderCode = ""

	updatedBattle, _ := json.Marshal(rb)
	msg := createSocketEvent("battle_revised", string(updatedBattle), "")

	return msg, nil, false
}

// Delete handles deleting the battle
func (b *Service) Delete(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	err := b.db.DeleteBattle(BattleID)
	if err != nil {
		return nil, err, false
	}
	msg := createSocketEvent("battle_conceded", "", "")

	return msg, nil, false
}

// PlanAdd adds a new plan to the battle
func (b *Service) PlanAdd(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	var p struct {
		Name               string `json:"planName"`
		Type               string `json:"type"`
		ReferenceId        string `json:"referenceId"`
		Link               string `json:"link"`
		Description        string `json:"description"`
		AcceptanceCriteria string `json:"acceptanceCriteria"`
	}
	json.Unmarshal([]byte(EventValue), &p)

	plans, err := b.db.CreatePlan(BattleID, p.Name, p.Type, p.ReferenceId, p.Link, p.Description, p.AcceptanceCriteria)
	if err != nil {
		return nil, err, false
	}
	updatedPlans, _ := json.Marshal(plans)
	msg := createSocketEvent("plan_added", string(updatedPlans), "")

	return msg, nil, false
}

// PlanRevise handles editing a battle plan
func (b *Service) PlanRevise(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	var p struct {
		Id                 string `json:"planId"`
		Name               string `json:"planName"`
		Type               string `json:"type"`
		ReferenceId        string `json:"referenceId"`
		Link               string `json:"link"`
		Description        string `json:"description"`
		AcceptanceCriteria string `json:"acceptanceCriteria"`
	}
	json.Unmarshal([]byte(EventValue), &p)

	plans, err := b.db.RevisePlan(BattleID, p.Id, p.Name, p.Type, p.ReferenceId, p.Link, p.Description, p.AcceptanceCriteria)
	if err != nil {
		return nil, err, false
	}
	updatedPlans, _ := json.Marshal(plans)
	msg := createSocketEvent("plan_revised", string(updatedPlans), "")

	return msg, nil, false
}

// PlanDelete handles deleting a plan
func (b *Service) PlanDelete(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	plans, err := b.db.BurnPlan(BattleID, EventValue)
	if err != nil {
		return nil, err, false
	}
	updatedPlans, _ := json.Marshal(plans)
	msg := createSocketEvent("plan_burned", string(updatedPlans), "")

	return msg, nil, false
}

// PlanActivate handles activating a plan for voting
func (b *Service) PlanActivate(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	plans, err := b.db.ActivatePlanVoting(BattleID, EventValue)
	if err != nil {
		return nil, err, false
	}
	updatedPlans, _ := json.Marshal(plans)
	msg := createSocketEvent("plan_activated", string(updatedPlans), "")

	return msg, nil, false
}

// PlanSkip handles skipping a plan voting
func (b *Service) PlanSkip(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	plans, err := b.db.SkipPlan(BattleID, EventValue)
	if err != nil {
		return nil, err, false
	}
	updatedPlans, _ := json.Marshal(plans)
	msg := createSocketEvent("plan_skipped", string(updatedPlans), "")

	return msg, nil, false
}

// PlanFinalize handles setting a plan point value
func (b *Service) PlanFinalize(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	var p struct {
		Id     string `json:"planId"`
		Points string `json:"planPoints"`
	}
	json.Unmarshal([]byte(EventValue), &p)

	plans, err := b.db.FinalizePlan(BattleID, p.Id, p.Points)
	if err != nil {
		return nil, err, false
	}
	updatedPlans, _ := json.Marshal(plans)
	msg := createSocketEvent("plan_finalized", string(updatedPlans), "")

	return msg, nil, false
}

// Abandon handles setting abandoned true so battle doesn't show up in users battle list, then leaves battle
func (b *Service) Abandon(BattleID string, UserID string, EventValue string) ([]byte, error, bool) {
	b.db.AbandonBattle(BattleID, UserID)

	return nil, errors.New("ABANDONED_BATTLE"), true
}

// socketEvent is the event structure used for socket messages
type socketEvent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
	User  string `json:"warriorId"`
}

func createSocketEvent(Type string, Value string, User string) []byte {
	newEvent := &socketEvent{
		Type:  Type,
		Value: Value,
		User:  User,
	}

	event, _ := json.Marshal(newEvent)

	return event
}
