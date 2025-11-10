package consumer_bet

import (
	"fmt"
	"megin/app"
	"megin/app/api/service/es/doc"
	"megin/app/consumer"
	"megin/system"
	"testing"
)

func TestBetMessage(t *testing.T) {
	str := `[{
	"app_platform": "VI",
	"timezone": "Asia\/Bangkok",
	"event": "onBinlog",
	"game_order": {
		"id": 455070,
		"order_no": "FC20220616115903834RRZ8",
		"third_order_no": "62aaaa7b4721f00107b0c796",
		"offer_id": "62aaaa7b4721f00107b0c796",
		"uid": 2209,
		"pid": 21111,
		"game_id": 21003,
		"room_id": "21003",
		"room_name": "21003",
		"game_cid": 21003,
		"event_end": null,
		"event_start": null,
		"bet_at": "2022-06-16 11:59:03",
		"bet_type": "fish",
		"bet": "",
		"bet_result": 2,
		"bet_amount": "600.000000",
		"currency": "vndk",
		"game_currency": "vndk",
		"game_bet_amount": "600.000000",
		"reward_amount": "0.000000",
		"profit": "-600.000000",
		"game_profit": "0.000000",
		"bet_odds": null,
		"odds_style": "",
		"valid_amount": "600.000000",
		"state": 1,
		"remark": "",
		"settled_at": "2022-06-16 11:59:03",
		"rewater": "0.9000",
		"rewater_at": "2022-06-16 11:59:06",
		"is_valid": 1,
		"is_temp": 0,
		"created_at": "2022-06-16 11:59:03",
		"updated_at": "2023-03-10 13:45:15",
		"deleted_at": null,
		"rewater_rate": "0.15",
		"points": "0.000000",
		"debit_amount": "0.000000",
		"app_platform": "VI",
		"bet_content": "",
		"is_world_cup": 2,
		"bet_at_unix": 0,
		"settled_at_unix": 3
	},
	"user_info": {
		"username": "lan001010101",
		"realname": "LAN",
		"ip": "139.162.111.165",
		"source_code": "265Q1660",
		"account_type": "人人代理",
		"agent_id": 415,
		"code": "99W61430",
		"last_login_time": "2022-09-19T11:07:14.000000Z",
		"vip": 0,
		"game_user_name": "lan0010101_VI0034"
	},
	"game_info": {
		"game_slug": "fc",
		"game_title": "FC TRÒ CHƠI",
		"cate_id": 20014,
		"child_game_title": "",
		"child_game_slug": ""
	}
}]`
	var dataList []doc.GameOrderData
	list, err := consumer.BatchHandleGameOrders(dataList, []byte(str))

	err = consumer.SaveGameOrder(list)
	fmt.Println(list, err)

}

func TestWalletMessage(t *testing.T) {
	str := `{
	"before": null,
	"after": {
		"id": 10350030,
		"uid": 340802,
		"platform_id": 10024,
		"order_no": "EM202303101731557907P8P",
		"platform_order_no": "678259712800298631",
		"type": 22,
		"amount": 1000.0,
		"before_balance": 6511.351,
		"balance": 7511.351,
		"create_time": 1678469515000,
		"remark": "",
		"currency": "vndk",
		"msg": "",
		"platforms": null,
		"vip": 0,
		"wash_code": 0.0,
		"promo_id": 0,
		"send_time": null,
		"agent_name": "1E834967",
		"bet_msg": "",
		"account_type": "人人代理",
		"app_platform": "VI"
	},
	"source": {
		"version": "1.6.4.Final",
		"connector": "mysql",
		"name": "mysql_binlog_source",
		"ts_ms": 1678444315000,
		"snapshot": "false",
		"db": "yuanbao",
		"sequence": null,
		"table": "wallet",
		"server_id": 1583928901,
		"gtid": null,
		"file": "mysql-bin-changelog.164401",
		"pos": 6079196,
		"row": 0,
		"thread": null,
		"query": null
	},
	"op": "c",
	"ts_ms": 1678444315836,
	"transaction": null
}`

	var dataList = make([]doc.WalletChange, 2)
	list, err := consumer.BatchHandleWalletChange(dataList, 0, []byte(str), consumer.TopicWalletChangeBinlogVN)
	fmt.Println(list, err)
}

func TestMain(m *testing.M) {
	system.ServerInit("../../resources/config-dev.yaml", app.OnAppInitialize)
	m.Run()
}
