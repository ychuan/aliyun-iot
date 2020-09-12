package sign

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/thinkgos/aliyun-iot/infra"
)

const (
	testProductKey   = "a1QR3GD1Db3"
	testDeviceName   = "dynamic"
	testDeviceSecret = "632155d72f983dfe24b43b30c2ed9b2f"
)

func TestMQTTSign(t *testing.T) {
	t.Run("MQTT sign almost all", func(t *testing.T) {
		ms := New(
			WithSignMethod("hmacsha256"),
			WithExtRRPC(),
			WithSDKVersion("SDK-Golang-v1.13.3"),
			WithExtParamsKV("testKey", "testValue"),
		)
		signout, err := ms.Generate(
			&infra.MetaInfo{
				ProductKey:   testProductKey,
				DeviceName:   testDeviceName,
				DeviceSecret: testDeviceSecret,
			},
			infra.CloudRegionDomain{
				Region: infra.CloudRegionShangHai,
			})

		require.NoError(t, err)
		t.Logf("%+v", signout)
	})

	t.Run("MQTT sign custom cloud region", func(t *testing.T) {
		ms := New(WithSignMethod("hmacsha1"))
		signout, err := ms.Generate(
			&infra.MetaInfo{
				ProductKey:   testProductKey,
				DeviceName:   testDeviceName,
				DeviceSecret: testDeviceSecret,
			},
			infra.CloudRegionDomain{
				Region:       infra.CloudRegionCustom,
				CustomDomain: "iot.custom.com",
			})
		require.NoError(t, err)
		t.Logf("%+v", signout)
	})

	t.Run("MQTT sign empty custom cloud region", func(t *testing.T) {
		ms := New()
		_, err := ms.Generate(&infra.MetaInfo{
			ProductKey:   testProductKey,
			DeviceName:   testDeviceName,
			DeviceSecret: testDeviceSecret,
		}, infra.CloudRegionDomain{
			Region:       infra.CloudRegionCustom,
			CustomDomain: "",
		})
		require.Error(t, err)
	})
}
