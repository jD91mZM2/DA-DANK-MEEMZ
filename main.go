package main;

import (
	"os"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"encoding/binary"
	"io"
	"strings"
	"time"
	"math/rand"
	"io/ioutil"
	"path/filepath"
	"encoding/json"
)

const DIRNAME = "Dank";
var sounds = make(map[string][][]byte, 0);
var images map[string]string;

var statuses = []string{
	"hidden object games",
	"Oh... Sir!",
	"Minecraft 1.0 ALPHA",
	"with your mother",
	"something",
	"something else",
	"bored",
	"dead"}

type Settings struct{
	vc *discordgo.VoiceConnection
	playing bool
	commander string
}
var settings = make(map[string]*Settings);

func main(){
	args := os.Args[1:];

	if(len(args) < 1){
		fmt.Println("No token provided!");
		return;
	}
	token := args[0];

	fmt.Println("Loading...");

	err := os.MkdirAll(DIRNAME, 0755);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "I THINK THERE WAS ERROR", err);
		return;
	}
	files, err := ioutil.ReadDir(DIRNAME);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "I THINK THERE WAS ERROR", err);
		return;
	}
	for _, file := range files{
		if(file.IsDir()){
			continue;
		}
		name := file.Name();
		if(!strings.HasSuffix(name, ".dca")){
			continue;
		}

		bytes := make([][]byte, 0);
		err = load(name, &bytes);
		if(err != nil){
			continue;
		}
		
		name = strings.ToLower(strings.TrimSuffix(name, ".dca"));
		sounds[name] = bytes;
	}

	data, err := ioutil.ReadFile("Dank/images.json");
	if(err != nil){
		fmt.Fprintln(os.Stderr, "DAT images.json FILE EZ POOF!", err);
		images = make(map[string]string, 0);
	} else {
		err = json.Unmarshal(data, &images);
		if(err != nil){
			fmt.Fprintln(os.Stderr, "DO U EVEN JSON, BRUH?", err);
			images = make(map[string]string, 0);
		}
	}

	fmt.Println("Starting...");
	d, err := discordgo.New("Bot " + token);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "I THINK THERE WAS ERROR", err);
		return;
	}
	d.AddHandler(ready);
	d.AddHandler(messageCreate);
	d.AddHandler(messageUpdate);
	err = d.Open();

	if(err != nil){
		fmt.Fprintln(os.Stderr, "I THINK THERE WAS ERROR", err);
		return;
	}
	fmt.Println("Started!");

	<-make(chan struct{});
}

func load(file string, buffer *[][]byte) error{
	f, err := os.Open(filepath.Join(DIRNAME, file));
	if err != nil {
		fmt.Fprintln(os.Stderr, "FILE WAS WEIRD IDK:", err);
		return err;
	}

	var length int16;
	for {
		err := binary.Read(f, binary.LittleEndian, &length);

		if(err == io.EOF || err == io.ErrUnexpectedEOF){
			break;
		} else if(err != nil){
			fmt.Fprintln(os.Stderr, "IDK WAT U DID BUT WEL DONN NOOB,", err);
			return err;
		}

		buf := make([]byte, length);
		err = binary.Read(f, binary.LittleEndian, &buf);
		if(err != nil){
			fmt.Fprintln(os.Stderr, "IDK WAT U DID BUT WEL DONN NOOB,", err);
			return err;
		}

		*buffer = append(*buffer, buf);
	}
	return nil;
}

func play(buffer [][]byte, session *discordgo.Session, guild, channel string, s *Settings){
	s.playing = true;
	var err error;
	s.vc, err = session.ChannelVoiceJoin(guild, channel, false, true);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "LEL FREGGIN NOOB,", err);
		s.playing = false;
		return;
	}

	err = s.vc.Speaking(true);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "I HAS TO TYPE ALL DESE ERROR MESSAGES " +
		"HAVE SOME SYMPATHY PLS", err);
		s.playing = false;
		return;
	}

	for _, buf := range buffer {
		if s.vc == nil { return; }
		s.vc.OpusSend <- buf;
	}

	err = s.vc.Speaking(false);
	if(err != nil){
		fmt.Fprintln(os.Stderr, err);
	}
	err = s.vc.Disconnect();
	if(err != nil){
		fmt.Fprintln(os.Stderr, err);
	}
	s.playing = false;
}

func messageCreate(session *discordgo.Session, event *discordgo.MessageCreate){
	message(session, event.Message)
}
func messageUpdate(session *discordgo.Session, event *discordgo.MessageUpdate){
	message(session, event.Message)
}
func message(session *discordgo.Session, event *discordgo.Message){
	msg := strings.ToLower(strings.TrimSpace(event.Content));
	author := event.Author;

	channel, err := session.State.Channel(event.ChannelID);
	if(err != nil){ fmt.Fprintln(os.Stderr, err); return; }

	guild, err := session.State.Guild(channel.GuildID);
	if(err != nil){ fmt.Fprintln(os.Stderr, err); return; }

	s := settings[guild.ID];
	if(s == nil){
		s = &Settings{};
		settings[guild.ID] = s;
	}

	if(s.commander != "" && s.commander != author.ID){
		return;
	}

	buffer, ok := sounds[msg];
	if(ok){
		if(!s.playing){
			for _, state := range guild.VoiceStates{
				if state.UserID == event.Author.ID{
					go react(session, event);
					play(buffer, session, guild.ID, state.ChannelID, s);
					return;
				}
			}
		}
		return;
	}

	for keyword, url := range images{
		if(strings.Contains(msg, strings.ToLower(keyword))){
			go react(session, event);
			_, err = session.ChannelMessageSendEmbed(event.ChannelID,
				&discordgo.MessageEmbed{
					Image: &discordgo.MessageEmbedImage{
						URL: url,
					},
				});
			return;
		}
	}

	switch(msg){
		case "thx":
			if(s.vc != nil && s.playing){
				err := s.vc.Speaking(false);
				if(err != nil){ fmt.Fprintln(os.Stderr, err); }

				err = s.vc.Disconnect();
				if(err != nil){ fmt.Fprintln(os.Stderr, err); }

				s.playing = false;
			}
		case "listen only to me plz":
			s.commander = author.ID;
		case "every1 owns u stopad robot":
			s.commander = "";
	}
}

func ready(session *discordgo.Session, event *discordgo.Ready){
	c := time.Tick(time.Second * 5);

	for _ = range c{
		err := session.UpdateStatus(0, statuses[rand.Intn(len(statuses))]);
		if(err != nil){ fmt.Fprintln(os.Stderr, err); return; }
	}
}

func react(session *discordgo.Session, event *discordgo.Message){
	err := session.MessageReactionAdd(event.ChannelID, event.ID, "👌");
	if(err != nil){
		fmt.Fprintln(os.Stderr, "Couldn't react,", err);
		return;
	}
	err = session.MessageReactionAdd(event.ChannelID, event.ID, "😂");
	if(err != nil){
		fmt.Fprintln(os.Stderr, "Couldn't react,", err);
	}
}
