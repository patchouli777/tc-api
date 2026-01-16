package setup

import (
	"fmt"
	"log"
	"math"
	"math/rand/v2"
	"os"
)

var tags = []string{
	"fun",
	"chill",
	"competitive",
	"live",
	"shooter",
	"war",
	"fantasy",
	"arts",
	"sport",
	"indie",
	"real life",
	"simulator",
	"adventure",
	"roleplay",
	"strategic",
	"platformer",
	"tutorial",
}

type setupSubTiers struct {
	Name  string
	Price float32
}

var subTiers = []setupSubTiers{
	{Name: "tier1", Price: 4.99},
	{Name: "tier2", Price: 9.99},
	{Name: "tier3", Price: 14.99}}

type setupCategory struct {
	Name      string
	Thumbnail string
	Link      string
}

var categories = []setupCategory{
	{Name: "world of warcraft", Thumbnail: "images/world-of-warcraft.png", Link: "world-of-warcraft"},
	{Name: "apex", Thumbnail: "images/apex.jpg", Link: "apex"},
	{Name: "dota 2", Thumbnail: "images/dota-2.jpg", Link: "dota-2"},
	{Name: "worlds beyond", Thumbnail: "images/worlds-beyond.jpg", Link: "worlds-beyond"},
	{Name: "art", Thumbnail: "images/art.jpg", Link: "art"},
	{Name: "black desert", Thumbnail: "images/black-desert.jpg", Link: "black-desert"},
	{Name: "counter strike 2", Thumbnail: "images/counter-strike-2.jpg", Link: "counter-strike-2"},
	{Name: "dead by daylight", Thumbnail: "images/dead-by-daylight.jpg", Link: "dead-by-daylight"},
	{Name: "death stranding 2", Thumbnail: "images/death-stranding-2.jpg", Link: "death-stranding-2"},
	{Name: "dune awakening", Thumbnail: "images/dune-awakening.jpg", Link: "dune-awakening"},
	{Name: "elden ring nightreign", Thumbnail: "images/elden-ring-nightreign.jpg", Link: "elden-ring-nightreign"},
	{Name: "escape from tarkov", Thumbnail: "images/escape-from-tarkov.jpg", Link: "escape-from-tarkov"},
	{Name: "fifa", Thumbnail: "images/fifa.jpg", Link: "fifa"},
	{Name: "fortnite", Thumbnail: "images/fortnite-battleroyale.jpg", Link: "fortnite"},
	{Name: "grand theft auto 5", Thumbnail: "images/grand-theft-auto-5.jpg", Link: "grand-theft-auto-5"},
	{Name: "hearthstone", Thumbnail: "images/hearthstone.png", Link: "hearthstone"},
	{Name: "irl", Thumbnail: "images/irl.jpg", Link: "irl"},
	{Name: "just chatting", Thumbnail: "images/just-chatting.jpg", Link: "just-chatting"},
	{Name: "league of legends", Thumbnail: "images/league-of-legends.jpg", Link: "league-of-legends"},
	{Name: "lycans", Thumbnail: "images/lycans.jpg", Link: "lycans"},
	{Name: "maplestory worlds", Thumbnail: "images/maplestory-worlds.jpg", Link: "maplestory-worlds"},
	{Name: "marvel rivals", Thumbnail: "images/marvel-rivals.jpg", Link: "marvel-rivals"},
	{Name: "minecraft", Thumbnail: "images/minecraft.jpg", Link: "minecraft"},
	{Name: "oldschool runescape", Thumbnail: "images/oldschool-runescape.jpg", Link: "oldschool-runescape"},
	{Name: "overwatch 2", Thumbnail: "images/overwatch-2.jpg", Link: "overwatch-2"},
	{Name: "path of exile 2", Thumbnail: "images/path-of-exile-2.jpg", Link: "path-of-exile-2"},
	{Name: "peak", Thumbnail: "images/peak.jpg", Link: "peak"},
	{Name: "phasmophobia", Thumbnail: "images/phasmophobia.jpg", Link: "phasmophobia"},
	{Name: "podscasts", Thumbnail: "images/podscasts.png", Link: "podscasts"},
	{Name: "pubg", Thumbnail: "images/pubg.png", Link: "pubg"},
	{Name: "rainbow6 siege", Thumbnail: "images/rainbow6-siegex.jpg", Link: "rainbow6-siege"},
	{Name: "rematch", Thumbnail: "images/rematch.jpg", Link: "rematch"},
	{Name: "rust", Thumbnail: "images/rust.jpg", Link: "rust"},
	{Name: "scum", Thumbnail: "images/scum.jpg", Link: "scum"},
	{Name: "sleep", Thumbnail: "images/sleep.jpg", Link: "sleep"},
	{Name: "sports", Thumbnail: "images/sports.png", Link: "sports"},
	{Name: "street fighter", Thumbnail: "images/street-fighter.jpg", Link: "street-fighter"},
	{Name: "teamfight tactics", Thumbnail: "images/teamfight-tactics.jpg", Link: "teamfight-tactics"},
	{Name: "tibia", Thumbnail: "images/tibia.jpg", Link: "tibia"},
	{Name: "valorant", Thumbnail: "images/valorant.jpg", Link: "valorant"},
	{Name: "warzone", Thumbnail: "images/warzone.png", Link: "warzone"},
	{Name: "world of tanks", Thumbnail: "images/world-of-tanks.jpg", Link: "world-of-tanks"}}

type setupUser struct {
	Name        string
	Password    string
	Avatar      string
	Description string
	Links       []string
}

func usersGet(count int) []setupUser {
	entries, err := os.ReadDir("./static/avatars")
	if err != nil {
		log.Print("couldn't load avatars")
	}

	avatars := make([]string, len(entries))
	for i, entry := range entries {
		avatars[i] = "avatars\\" + entry.Name()
	}

	if len(avatars) == 0 {
		avatars = append(avatars, "couln't load avatars")
	}

	avsLen := len(avatars)
	users := make([]setupUser, count)
	for i := range count {
		avId := rand.IntN(avsLen)

		users[i] = setupUser{
			Name:        fmt.Sprintf("user%d", i),
			Password:    fmt.Sprintf("password%d", i),
			Avatar:      avatars[avId],
			Description: fmt.Sprintf("description%d", i),
			Links:       []string{"instagram", "telegram"},
		}
	}

	return users
}

const streamsCountRatio = 0.25 / float64(1)
const usersCount = 100
const categoriesCount = 5
const followCountRatio = 1 / float64(2)

var followCount = int(math.Ceil(usersCount * followCountRatio))
var streamsCount = int(math.Ceil(usersCount * streamsCountRatio))
var users = usersGet(usersCount)
