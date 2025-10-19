package notification

type createGroupRequest struct {
	Name string `json:"name"`
}

type createChannelRequest struct {
	NotificationChannelGroupName string `json:"notificationChannelGroupName"`
	DestinationName              string `json:"destinationName"`
	Settings                     string `json:"settings"`
}

type updateChannelRequest struct {
	DestinationName string `json:"destinationName"`
	Settings        string `json:"settings"`
}
