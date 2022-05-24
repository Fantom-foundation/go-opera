package opera

import "encoding/json"

func UpdateRules(src Rules, diff []byte) (res Rules, err error) {
	changed := src.Copy()
	err = json.Unmarshal(diff, &changed)
	if err != nil {
		return src, err
	}
	// protect readonly fields
	res = changed
	res.NetworkID = src.NetworkID
	res.Name = src.Name
	// don't allow to revert an activated upgrade
	res.Upgrades.Berlin = res.Upgrades.Berlin || src.Upgrades.Berlin
	res.Upgrades.London = res.Upgrades.London || src.Upgrades.London
	res.Upgrades.Llr = res.Upgrades.Llr || src.Upgrades.Llr
	return
}
