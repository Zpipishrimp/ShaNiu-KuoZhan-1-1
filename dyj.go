package jd_cookie

func init() {
	// //帮助作者助力大赢家，如介意请删除本扩展。
	// go func() {
	// 	time.Sleep(time.Second)
	// 	date := time.Now().Format("2006-01-02")
	// 	defer jd_cookie.Set("dyj_help", date)
	// 	if jd_cookie.Get("dyj_help") != date {
	// 		envs, err := qinglong.GetEnvs("JD_COOKIE")
	// 		if err != nil {
	// 			return
	// 		}
	// 		s := 1
	// 		for i := 0; i < len(envs); i++ {
	// 			if envs[i].Status == 0 {
	// 				req := httplib.Get("https://api.m.jd.com/?functionId=openRedEnvelopeInteract&body=" + `{"linkId":"yMVR-_QKRd2Mq27xguJG-w","redEnvelopeId":"48855bbdac3443bfbbc5d6cd34c9ef4566211633281100281","inviter":"WaHpHMRJdh28jqa9WTwWl3LXebXXp5CBbOkCxVi4jTg","helpType":"` + fmt.Sprint(s) + `"}` + "&t=" + fmt.Sprint(time.Now().Unix()) + "&appid=activities_platform&clientVersion=3.5.6")
	// 				req.Header("Cookie", envs[i].Value)
	// 				req.Header("Accept", "*/*")
	// 				req.Header("Connection", "keep-alive")
	// 				req.Header("Accept-Encoding", "gzip, deflate, br")
	// 				req.Header("Host", "api.m.jd.com")
	// 				req.Header("Origin", "https://wbbny.m.jd.com")
	// 				data, _ := req.String()
	// 				if strings.Contains(data, "提现") {
	// 					if s == 1 {
	// 						s = 2
	// 					} else {
	// 						break
	// 					}
	// 				}
	// 			}
	// 		}

	// 	}
	// }()
}
