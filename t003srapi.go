/*!
Copyright © 2022 chouette.21.00@gmail.com
Released under the MIT license
https://opensource.org/licenses/mit-license.php

Ver. 0.0.0

*/
package main

import (
	"log"
	"os"
	"time"

	"github.com/Chouette2100/exsrapi"
	"github.com/Chouette2100/srapi"
)

/*
	配信中ルームの一覧を取得して、星集め/種集めの対象とするルームのリストを作る

	Githubのサンプルプログラムの入手方法・利用方法については次の記事をご参照ください。

		【Windows】Githubにあるサンプルプログラムの実行方法
			https://zenn.dev/chouette2100/books/d8c28f8ff426b7/viewer/e27fc9

		【Unix/Linux】Githubにあるサンプルプログラムの実行方法
			https://zenn.dev/chouette2100/books/d8c28f8ff426b7/viewer/220e38


	$ cd ~/go/src/t003srapi
	$ vi t003srapi.go
	$ go mod init
	$ go mod tidy
	$ go build t001srapi.go
	$ cat config.yml
	category: Official				<== "Free"|"Official"
	rvlfilename: rvl.txt
	exclfilename: excl.txt
	aplmin: 240
	maxnoroom: 20
	$ cat rvl.txt					<== 訪問済みルームリストのサンプル（星集め・種集めにともなって作成・更新されるファイル）
	"2022/08/11 11:01:04 +0900 JST"	333333
	"2022/08/11 11:07:18 +0900 JST"	222222
	"2022/08/11 12:03:18 +0900 JST"	111111
	$ cat excl.txt					<== 除外リストのサンプル（事前に作成しておくファイル）
	100000	(応援ルーム)xxxx0
	100001	(応援ルーム)xxxx1
	100002	(関わりたくないルーム)xxxx2
	100003	(関わりたくないルーム)xxxx3
	100004	(星がもらえないルール)xxxx4
	$ ./t003srapi config.yml

*/

type Config struct {
	SR_acct     string //	SHOWROOMのアカウント名
	SR_pswd     string //	SHOWROOMのパスワード
	Category    string //	カテゴリー名
	Aplmin      int    //	訪問ルームリストの有効時間(分)
	Maxnoroom   int    //	訪問候補ルームリストの最大長
	Rvlfilename string //	訪問済みルームリストファイル名
	Exclfilename string	//	除外ルームリストファイル名
}

func main() {

	//	ログファイルを設定する。
	logfile := exsrapi.CreateLogfile("", "")
	defer logfile.Close()

	if len(os.Args) != 2 {
		//      引数が足りない(設定ファイル名がない)
		log.Printf("usage:  %s NameOfConfigFile\n", os.Args[0])
		return
	}

	//	設定ファイルを読み込む
	var config Config
	err := exsrapi.LoadConfig(os.Args[1], &config)
	if err != nil {
		log.Printf("exsrapi.LoadConfig: %s\n", err.Error())
		return
	}

	//	cookiejarがセットされたHTTPクライアントを作る
	client, jar, err := exsrapi.CreateNewClient(config.SR_acct)
	if err != nil {
		log.Printf("CreateNewClient: %s\n", err.Error())
		return
	}
	//	すべての処理が終了したらcookiejarを保存する。
	defer jar.Save()

	//	配信しているルームの一覧を取得する
	roomlives, err := srapi.ApiLiveOnlives(client)
	if err != nil {
		log.Printf("ApiLiveOnlives(): %s\n", err.Error())
		return
	}
	log.Printf("*****************************************************************\n")
	log.Printf("配信中ルーム数\n")
	log.Printf("\n")
	log.Printf("　ジャンル数= %d\n", len(roomlives.Onlives))
	log.Printf("\n")
	log.Printf("　ルーム数　ジャンル　ジャンル名\n")
	for _, roomlive := range roomlives.Onlives {
		log.Printf("%10d%10d  %s\n", len(roomlive.Lives), roomlive.Genre_id, roomlive.Genre_name)
	}
	log.Printf("\n")

	roomvisit := new(exsrapi.RoomVisit) //	訪問ルームリスト
	roomvisit.Roomvisit = make(map[int]time.Time)

	excllist := exsrapi.ExclList{}    //	除外ルームリスト
	err = excllist.Read(config.Category, config.Exclfilename)     //	除外ルームリストを読み込む
	if err != nil {
		log.Printf("excllist.Read(): %s\n", err.Error())
		return
	}

	//	訪問ルームリストファイルからすでに星集め、種集めのために訪問したリストを読み込む
	err = roomvisit.Restore(config.Category, config.Rvlfilename, config.Aplmin)
	if err != nil {
		log.Printf("RestoreRVL(): %s\n", err.Error())
		return
	}
	defer roomvisit.Save() //	訪問したリストを保存する。本来星集め、種集めが終わったあと行う処理。

	//	星集め/種集めの対象とするルームのリストを作る
	lives, err := exsrapi.MkRoomsForStarCollec(client, config.Category, config.Aplmin, config.Maxnoroom, &excllist, &roomvisit.Roomvisit)
	if err != nil {
		log.Printf("MkRoomsForStarCollec(): %s\n", err.Error())
		return
	}

	log.Printf("%-9s*** %d rooms Sorted.\n", config.Category, len(*lives))
	for _, live := range *lives {
		log.Printf("%-9s  %-12d%s %s\n", config.Category, live.Room_id, time.Unix(live.Started_at, 0).Format("01-02 15:04:05"), live.Main_name)
	}

}
