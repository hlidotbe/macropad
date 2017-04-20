package auxilium

import (
	"fmt"
	"net/http"
)

// TimeTrackService provides a way to manipulate timetracks
type TimeTrackService struct {
	client *Client
}

// Create a new running TimeTrack
func (t *TimeTrackService) Create(time *TimeTrack) (int, *http.Response, error) {
	opt := timeTrackRequest{TimeTrack: time}
	req, err := t.client.NewRequest("POST", "time_tracks", opt)
	if err != nil {
		return 0, nil, err
	}
	resp, err := t.client.Do(req, time)
	if err != nil {
		return 0, resp, err
	}
	return time.Id, resp, nil
}

// Update an existing TimeTrack
func (t *TimeTrackService) Update(time *TimeTrack) (*http.Response, error) {
	opt := timeTrackRequest{TimeTrack: time}
	req, err := t.client.NewRequest("PUT", fmt.Sprintf("time_tracks/%d", time.Id), opt)
	if err != nil {
		return nil, err
	}
	resp, err := t.client.Do(req, time)
	if err != nil {
		return resp, err
	}
	return resp, nil
}

// Show an existing TimeTrack
func (t *TimeTrackService) Show(id int) (*TimeTrack, *http.Response, error) {
	req, err := t.client.NewRequest("GET", fmt.Sprintf("time_tracks/%d", id), nil)
	if err != nil {
		return nil, nil, err
	}
	time := new(TimeTrack)
	resp, err := t.client.Do(req, time)
	if err != nil {
		return nil, resp, err
	}
	return time, resp, nil
}

// A TimeTrack represents a unit of work
type TimeTrack struct {
	Id        int      `json:"id,omitempty"`
	ProjectId int      `json:"project_id,omitempty"`
	Started   string   `json:"started,omitempty"`
	Duration  int      `json:"duration,omitempty"`
	Notes     int      `json:"notes,omitempty"`
	Billable  bool     `json:"billable,omitempty"`
	Profile   string   `json:"profile,omitempty"`
	Direction bool     `json:"direction,omitempty"`
	Status    string   `json:"status,omitempty"`
	Tickets   []int    `json:"tickets,omitempty"`
	Clients   []int    `json:"clients,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	CreatedAt string   `json:"created_at,omitempty"`
	UpdatedAt string   `json:"updated_at,omitempty"`
	UserId    int      `json:"user_id,omitempty"`
}

type timeTrackRequest struct {
	TimeTrack *TimeTrack `json:"time_track"`
}
