package main

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
)

const DIRNAME = "Dank";
var sounds = make(map[string][][]byte, 0);

const FEELSBADMAN = "https://openclipart.org/image/2400px/svg_to_png/222252/" +
"feels.png";

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
		fmt.Fprintln(os.Stderr, "I THINK THERE WAS ERROR ", err);
		return;
	}
	files, err := ioutil.ReadDir(DIRNAME);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "I THINK THERE WAS ERROR ", err);
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
		load(name, &bytes);
		
		name = strings.ToLower(strings.TrimSuffix(name, ".dca"));
		sounds[name] = bytes;
	}

	fmt.Println("Starting...");
	d, err := discordgo.New("Bot " + token);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "I THINK THERE WAS ERROR ", err);
		return;
	}
	d.AddHandler(ready);
	d.AddHandler(messageCreate);
	d.AddHandler(messageUpdate);
	err = d.Open();

	if(err != nil){
		fmt.Fprintln(os.Stderr, "I THINK THERE WAS ERROR ", err);
		return;
	}
	fmt.Println("Started!");

	<-make(chan struct{});
}

func load(file string, buffer *[][]byte){
	f, err := os.Open(filepath.Join(DIRNAME, file));
	if err != nil {
		fmt.Fprintln(os.Stderr, "FILE WAS WEIRD IDK: ", err);
		return;
	}

	var length int16;
	for {
		err := binary.Read(f, binary.LittleEndian, &length);

		if(err == io.EOF || err == io.ErrUnexpectedEOF){
			return;
		} else if(err != nil) {
			fmt.Fprintln(os.Stderr, "IDK WAT U DID BUT WEL DONN NOOB, ", err);
			return;
		}

		buf := make([]byte, length);
		binary.Read(f, binary.LittleEndian, &buf);

		*buffer = append(*buffer, buf);
	}
}

func play(buffer [][]byte, session *discordgo.Session, guild, channel string, s *Settings){
	s.playing = true;
	var err error;
	s.vc, err = session.ChannelVoiceJoin(guild, channel, false, true);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "LEL FREGGIN NOOB, ", err);
		s.playing = false;
		return;
	}

	err = s.vc.Speaking(true);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "I HAS TO TYPE ALL DESE ERROR MESSAGES " +
		"HAVE SOME SYMPATHY PLS ", err);
		return;
	}

	for _, buf := range buffer {
		if s.vc == nil { return; }
		s.vc.OpusSend <- buf;
	}

	err = s.vc.Speaking(false);
	if(err != nil){
		fmt.Fprintln(os.Stderr, err);
		return;
	}
	err = s.vc.Disconnect();
	if(err != nil){
		fmt.Fprintln(os.Stderr, err);
		return;
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

	var image string = "";
	var buffer [][]byte = nil;

	buffer2, ok := sounds[msg];
	if(ok){
		buffer = buffer2;
	} else {
		switch(msg){
			case "feelsbadman":
				image = FEELSBADMAN;
			case "thx":
				if(s.vc != nil){
					err := s.vc.Speaking(false);
					if(err != nil){ fmt.Fprintln(os.Stderr, err); return; }

					err = s.vc.Disconnect();
					if(err != nil){ fmt.Fprintln(os.Stderr, err); return; }

					s.playing = false;
				}
			case "listen only to me plz":
				s.commander = author.ID;
			case "every1 owns u stopad robot":
				s.commander = "";
		}
	}

	if(image != ""){
		_, err = session.ChannelMessageSendEmbed(event.ChannelID,
		&discordgo.MessageEmbed{
			Image: &discordgo.MessageEmbedImage{
				URL: image,
			},
		});
	} else if(buffer != nil && !s.playing){
		for _, state := range guild.VoiceStates{
			if state.UserID == event.Author.ID{
				play(buffer, session, guild.ID, state.ChannelID, s);
			}
		}
	}
}

func ready(session *discordgo.Session, event *discordgo.Ready){
	c := time.Tick(time.Second * 5);

	for _ = range c{
		err := session.UpdateStatus(0, statuses[rand.Intn(len(statuses))]);
		if(err != nil){ fmt.Fprintln(os.Stderr, err); return; }
	}
}
