package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/stdutil"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
	"time"
)

type Image struct {
	Keyword string
	Image   string
}

const DIRNAME = "Dank"

var sounds = make(map[string][][]byte, 0)
var images []*Image

var statuses = []string{
	"hidden object games",
	"Oh... Sir!",
	"Minecraft 1.0 ALPHA",
	"with your mother",
	"something",
	"something else",
	"bored",
	"dead"}

type Settings struct {
	playing   bool
	commander string
}

var settings = make(map[string]*Settings)

func main() {
	//stdutil.ShouldTrace = true;
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("No token provided!")
		return
	}
	token := args[0]

	fmt.Println("Loading...")

	err := os.MkdirAll(DIRNAME, 0755)
	if err != nil {
		stdutil.PrintErr("", err)
		return
	}
	files, err := ioutil.ReadDir(DIRNAME)
	if err != nil {
		stdutil.PrintErr("", err)
		return
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if !strings.HasSuffix(name, ".dca") {
			continue
		}

		bytes := make([][]byte, 0)
		err = load(name, &bytes)
		if err != nil {
			continue
		}

		name = strings.ToLower(strings.TrimSuffix(name, ".dca"))
		sounds[name] = bytes
	}

	data, err := ioutil.ReadFile("Dank/images.json")
	if err != nil {
		stdutil.PrintErr("", err)
	} else {
		var imagesMap map[string]string
		err = json.Unmarshal(data, &imagesMap)

		for key, val := range imagesMap {
			images = append(images, &Image{
				Keyword: key,
				Image:   val,
			})
		}
		sort.Slice(images, func(i, j int) bool {
			return len(images[i].Keyword) > len(images[j].Keyword)
		})
		if err != nil {
			stdutil.PrintErr("", err)
		}
	}

	fmt.Println("Starting...")
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		stdutil.PrintErr("", err)
		return
	}
	session.AddHandler(messageCreate)
	session.AddHandler(messageUpdate)
	err = session.Open()

	if err != nil {
		stdutil.PrintErr("", err)
		return
	}

	go func() {
		c := time.Tick(time.Second * 5)

		for _ = range c {
			err := session.UpdateStatus(0, statuses[rand.Intn(len(statuses))])
			if err != nil {
				stdutil.PrintErr("", err)
				return
			}
		}
	}()
	fmt.Println("Started!")

	interrupt := make(chan os.Signal, 2)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	<-interrupt
	fmt.Println("\nExiting")
	session.Close()
}

func load(file string, buffer *[][]byte) error {
	f, err := os.Open(filepath.Join(DIRNAME, file))
	defer f.Close()
	if err != nil {
		stdutil.PrintErr("", err)
		return err
	}

	var length int16
	for {
		err := binary.Read(f, binary.LittleEndian, &length)

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		} else if err != nil {
			stdutil.PrintErr("", err)
			return err
		}

		buf := make([]byte, length)
		err = binary.Read(f, binary.LittleEndian, &buf)
		if err != nil {
			stdutil.PrintErr("", err)
			return err
		}

		*buffer = append(*buffer, buf)
	}
	return nil
}

func play(buffer [][]byte, session *discordgo.Session, guild, channel string, s *Settings) {
	s.playing = true
	defer func() { s.playing = false }()
	vc, err := session.ChannelVoiceJoin(guild, channel, false, true)
	if err != nil {
		stdutil.PrintErr("", err)
		return
	}

	err = vc.Speaking(true)
	if err != nil {
		stdutil.PrintErr("", err)

		err = vc.Disconnect()
		if err != nil {
			stdutil.PrintErr("", err)
		}
		return
	}

	for _, buf := range buffer {
		if !s.playing {
			break
		}
		vc.OpusSend <- buf
	}

	err = vc.Speaking(false)
	if err != nil {
		stdutil.PrintErr("", err)
	}
	err = vc.Disconnect()
	if err != nil {
		stdutil.PrintErr("", err)
	}
}

func messageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	message(session, event.Message)
}
func messageUpdate(session *discordgo.Session, event *discordgo.MessageUpdate) {
	message(session, event.Message)
}
func message(session *discordgo.Session, event *discordgo.Message) {
	if event.Author == nil {
		return
	}
	msg := strings.ToLower(strings.TrimSpace(event.Content))
	author := event.Author

	if msg == "" {
		return
	}

	channel, err := session.Channel(event.ChannelID)
	if err != nil {
		stdutil.PrintErr("", err)
		return
	}

	if channel.IsPrivate {
		return
	}

	guild, err := session.Guild(channel.GuildID)
	if err != nil {
		stdutil.PrintErr("", err)
		return
	}

	s := settings[guild.ID]
	if s == nil {
		s = &Settings{}
		settings[guild.ID] = s
	}

	if msg == "meemz who ur master" {
		msg := ""
		if s.commander == "" {
			msg = "nobody dos idiot"
		} else if s.commander == author.ID {
			msg = "u do... idiot"
		} else {
			msg = "dat wuld b <@" + s.commander + ">"
		}
		_, err = session.ChannelMessageSend(event.ChannelID, msg)
		if err != nil {
			stdutil.PrintErr("", err)
		}
	}

	if guild.OwnerID == author.ID && strings.HasPrefix(msg, "meemz idfc ") {
		msg = strings.TrimPrefix(msg, "meemz idfc ")
	} else {
		if s.commander != "" && s.commander != author.ID {
			return
		}
	}

	buffer, ok := sounds[msg]
	if ok {
		if !s.playing {
			for _, state := range guild.VoiceStates {
				if state.UserID == event.Author.ID {
					go react(session, event)
					play(buffer, session, guild.ID, state.ChannelID, s)
					return
				}
			}
		}
		return
	}

	for _, image := range images {
		contains, err := regexp.MatchString("(?i)\\b"+regexp.QuoteMeta(image.Keyword)+"\\b", msg)
		if err != nil {
			stdutil.PrintErr("", err)
			return
		}
		if contains {
			go react(session, event)
			_, err = session.ChannelMessageSendEmbed(event.ChannelID,
				&discordgo.MessageEmbed{
					Image: &discordgo.MessageEmbedImage{
						URL: image.Image,
					},
				})
			if err != nil {
				stdutil.PrintErr("", err)
			}
			return
		}
	}

	switch msg {
	case "thx":
		s.playing = false
	case "listen only to me plz":
		s.commander = author.ID
		fmt.Println("In guild '" + guild.Name + "', the user '" + author.Username + "' took control.")
	case "every1 owns u stopad robot":
		s.commander = ""
		fmt.Println("In guild '" + guild.Name + "', the user '" + author.Username + "' returned the control to everyone.")
	case "plz list da stuff":
		strSounds := ""
		for name := range sounds {
			if strSounds != "" {
				strSounds += ", "
			}
			strSounds += "`" + name + "`"
		}

		strImages := ""
		for _, image := range images {
			if strImages != "" {
				strImages += ", "
			}
			strImages += "`" + image.Keyword + "`"
		}

		_, err := session.ChannelMessageSendEmbed(event.ChannelID, &discordgo.MessageEmbed{
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:  "DA SOUNDZ",
					Value: strSounds,
				},
				&discordgo.MessageEmbedField{
					Name:   "DA IMAGEZ",
					Value:  strImages,
					Inline: true,
				},
			},
		})
		if err != nil {
			stdutil.PrintErr("", err)
			return
		}
	}
}

func react(session *discordgo.Session, event *discordgo.Message) {
	err := session.MessageReactionAdd(event.ChannelID, event.ID, "👌")
	if err != nil {
		stdutil.PrintErr("", err)
		return
	}
	err = session.MessageReactionAdd(event.ChannelID, event.ID, "😂")
	if err != nil {
		stdutil.PrintErr("", err)
	}
}
