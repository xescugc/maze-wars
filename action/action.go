package action

type Action struct {
	Type Type `json:"type"`

	CameraMove CameraMovePayload `json:"camera_move,omitempty"`
}

type CameraMovePayload struct {
	X int `json:"x"`
	Y int `json:"y"`
}

func NewCameraMove(x, y int) *Action {
	return &Action{
		Type: CameraMove,
		CameraMove: CameraMovePayload{
			X: x,
			Y: y,
		},
	}
}
