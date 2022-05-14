package dto

//{
//     "enclosureLocation": null,
//     "id": "\\\\?\\SWD#ScDeviceEnum#1_ACS_ACR1252_1S_CL_Reader_PICC_0#{deebe6ad-9e01-47e2-a3b2-a66aa2c036c9}",
//     "isDefault": false,
//     "isEnabled": true,
//     "kind": 1,
//     "name": "ACS ACR1252 1S CL Reader PICC 0",
//     "pairing": {
//         "canPair": false,
//         "custom": {},
//         "isPaired": false,
//         "protectionLevel": 1
//     },
//     "properties": {
//         "System.ItemNameDisplay": "ACS ACR1252 1S CL Reader PICC 0",
//         "System.Devices.DeviceInstanceId": "SWD\\ScDeviceEnum\\1_ACS_ACR1252_1S_CL_Reader_PICC_0",
//         "System.Devices.Icon": "C:\\Windows\\System32\\DDORes.dll,-2068",
//         "System.Devices.GlyphIcon": "C:\\Windows\\System32\\DDORes.dll,-3001",
//         "System.Devices.InterfaceEnabled": true,
//         "System.Devices.IsDefault": false,
//         "System.Devices.PhysicalDeviceLocation": null,
//         "System.Devices.ContainerId": "00000000-0000-0000-ffff-ffffffffffff"
//     }
// }

type DeviceInformation struct {
	EnclosureLocation interface{}       `json:"enclosureLocation"`
	ID                string            `json:"id"`
	IsDefault         bool              `json:"isDefault"`
	IsEnabled         bool              `json:"isEnabled"`
	Kind              int               `json:"kind"`
	Name              string            `json:"name"`
	Pairing           *DevicePairing    `json:"pairing"`
	Properties        *DeviceProperties `json:"properties"`
}

type DevicePairing struct {
	CanPair         bool        `json:"canPair"`
	Custom          interface{} `json:"custom"`
	IsPaired        bool        `json:"isPaired"`
	ProtectionLevel int         `json:"rotectionLevel"`
}

type DeviceProperties struct {
	ItemNameDisplay  string `json:"System.ItemNameDisplay"`
	InterfaceEnabled bool   `json:"System.Devices.InterfaceEnabled"`
}

func NewDeviceInfo(name, id string, isEnabled bool) *DeviceInformation {
	dev := &DeviceInformation{}
	dev.IsEnabled = isEnabled
	dev.Name = name
	dev.ID = id
	dev.Kind = 1
	dev.Pairing = &DevicePairing{
		ProtectionLevel: 1,
	}
	dev.Properties = &DeviceProperties{
		ItemNameDisplay:  name,
		InterfaceEnabled: true,
	}

	return dev
}
