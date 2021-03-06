package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/legolord208/stdutil"
)

type image struct {
	Keyword string
	URL     string
}

const dir = "Dank"

var sounds = make(map[string][][]byte, 0)
var images []*image

var statuses = []string{
	"hidden object games",
	"Oh... Sir!",
	"Minecraft 1.0 ALPHA",
	"with your mother",
	"something",
	"something else",
	"bored",
	"dead"}

type settingsType struct {
	playing   bool
	commander string
}

var settings = make(map[string]*settingsType)
var settingsMutex sync.RWMutex

func main() {
	//stdutil.ShouldTrace = true;
	args := os.Args[1:]

	if len(args) < 1 {
		fmt.Println("No token provided!")
		return
	}
	token := args[0]

	fmt.Println("Loading...")

	err := os.Mkdir(dir, 0755)
	if err != nil && !os.IsExist(err) {
		stdutil.PrintErr("", err)
		return
	}
	files, err := ioutil.ReadDir(dir)
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

	file, err := os.Open("Dank/images.txt")
	if err != nil {
		stdutil.PrintErr("", err)
	} else {
		reader := bufio.NewScanner(file)
		for reader.Scan() {
			text := reader.Text()
			if text == "" {
				continue
			}
			parts := strings.SplitN(text, ", ", 2)
			if len(parts) != 2 {
				stdutil.PrintErr("Corrupt file or something", nil)
				continue
			}

			images = append(images, &image{
				Keyword: parts[0],
				URL:     parts[1],
			})
		}

		file.Close()

		if err := reader.Err(); err != nil {
			stdutil.PrintErr("Error reading file", err)
			return
		}

		sort.Slice(images, func(i int, i2 int) bool {
			return len(images[i].Keyword) > len(images[i2].Keyword)
		})
	}

	fmt.Println("Starting...")
	session, err := discordgo.New("Bot " + token)
	if err != nil {
		stdutil.PrintErr("", err)
		return
	}
	session.AddHandler(messageCreate)
	err = session.Open()

	if err != nil {
		stdutil.PrintErr("", err)
		return
	}

	go func() {
		c := time.Tick(time.Minute * 5)

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
	f, err := os.Open(filepath.Join(dir, file))
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

func play(buffer [][]byte, session *discordgo.Session, guild, channel string, s *settingsType) {
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
	if event.Author == nil {
		return
	}
	msg := strings.ToLower(strings.TrimSpace(event.Content))
	author := event.Author

	if msg == "" {
		return
	}

	channel, err := session.State.Channel(event.ChannelID)
	if err != nil {
		stdutil.PrintErr("", err)
		return
	}

	if channel.Type != discordgo.ChannelTypeGuildText {
		return
	}

	guild, err := session.State.Guild(channel.GuildID)
	if err != nil {
		stdutil.PrintErr("", err)
		return
	}

	settingsMutex.RLock()
	s := settings[guild.ID]
	settingsMutex.RUnlock()

	if s == nil {
		s = &settingsType{}

		settingsMutex.Lock()
		settings[guild.ID] = s
		settingsMutex.Unlock()
	}

	if msg == "meemz who ur master" {
		msg := ""
		if s.commander == "" {
			msg = "nobody is idiot"
		} else if s.commander == author.ID {
			msg = "u is... idiot"
		} else if s.commander == "-" {
			msg = "im no suppos 2 talk 2 u"
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
					go react(session, event.Message)
					play(buffer, session, guild.ID, state.ChannelID, s)
					return
				}
			}
		}
		return
	}

	for _, image := range images {
		if msg == image.Keyword {
			go react(session, event.Message)
			_, err = session.ChannelMessageSendEmbed(event.ChannelID,
				&discordgo.MessageEmbed{
					Image: &discordgo.MessageEmbedImage{
						URL: image.URL,
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
	case "meemz stfu":
		s.commander = "-"
		fmt.Println("In guild '" + guild.Name + "', the user '" + author.Username + "' disabled meemz.")
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
