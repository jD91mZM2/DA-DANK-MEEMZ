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
)

var JohnCena = make([][]byte, 0);
var Elevator = make([][]byte, 0);
var Rickroll = make([][]byte, 0);
var Letter = make([][]byte, 0);
var Cri = make([][]byte, 0);
var NumberHat = make([][]byte, 0);
var ExoticButters = make([][]byte, 0);

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
	
	load("John Cena", &JohnCena);
	load("Elevator", &Elevator);
	load("Rickroll", &Rickroll);
	load("Cri", &Cri);
	load("Letter", &Letter);
	load("NumberHat", &NumberHat);
	load("ExoticButters", &ExoticButters);
	
	d, _ := discordgo.New("Bot " + token);
	d.AddHandler(ready);
	d.AddHandler(messageCreate);
	d.Open();
	
	<-make(chan struct{});
}

func load(file string, buffer *[][]byte){
	f, err := os.Open("Dank/" + file + ".dca");
	if err != nil { fmt.Println("File not found: " + file); return; }
	
	var length int16;
	for {
		err := binary.Read(f, binary.LittleEndian, &length);
		
		if err == io.EOF || err == io.ErrUnexpectedEOF{
			return;
		}
		
		buf := make([]byte, length);
		binary.Read(f, binary.LittleEndian, &buf);
		
		*buffer = append(*buffer, buf);
	}
}

func play(buffer [][]byte, session *discordgo.Session, guild, channel string, s *Settings){
	s.playing = true;
	s.vc, _ = session.ChannelVoiceJoin(guild, channel, false, true);
	
	s.vc.Speaking(true);
	
	for _, buf := range buffer{
		if s.vc == nil { return; }
		s.vc.OpusSend <- buf;
	}
	
	s.vc.Speaking(false);
	s.vc.Disconnect();
	s.playing = false;
}

func messageCreate(session *discordgo.Session, event *discordgo.MessageCreate){
	msg := strings.ToLower(strings.TrimSpace(event.Content));
	author := event.Author;
	
	channel, _ := session.State.Channel(event.ChannelID);
	guild, _ := session.State.Guild(channel.GuildID);
	//member, _ := session.State.Member(guild.ID, author.ID);
	
	s := settings[guild.ID];
	if s == nil{
		s = &Settings{};
		settings[guild.ID] = s;
	}
	fmt.Println(*s);
	
	if(s.commander != "" && s.commander != author.ID){
		return;
	}
	
	var buffer [][]byte = nil;
	if msg == "john cena"{
		buffer = JohnCena;
	} else if msg == "waiting"{
		buffer = Elevator;
	} else if msg == "rickroll"{
		buffer = Rickroll;
	} else if msg == "cri"{
		buffer = Cri;
	} else if msg == "letter"{
		buffer = Letter;
	} else if msg == "numbr hat"{
		buffer = NumberHat;
	} else if msg == "exotic butters"{
		buffer = ExoticButters;
	} else if msg == "thx"{
		if s.vc != nil{
			s.vc.Speaking(false);
			s.vc.Disconnect();
			s.playing = false;
		}
	} else if msg == "listen only to me plz"{
		s.commander = author.ID;
	} else if msg == "every1 owns u stopad robot"{
		s.commander = "";
	} else if msg == "cler da chat plz"{
		messages, _ := session.ChannelMessages(event.ChannelID, 100, "", "");
		ids := make([]string, 0);
		for _, message := range messages{
			ids = append(ids, message.ID);
		}
		session.ChannelMessagesBulkDelete(event.ChannelID, ids);
	}
	
	if buffer != nil && !s.playing{
		for _, state := range guild.VoiceStates{
			if state.UserID == event.Author.ID{
				play(buffer, session, guild.ID, state.ChannelID, s);
			}
		}
	}
}

func ready(session *discordgo.Session, event *discordgo.Ready){
	ticker := time.NewTicker(time.Second * 5);
	go func(){
		for{
			<- ticker.C;
			session.UpdateStatus(0, statuses[rand.Intn(len(statuses))]);
		}
	}();
}
