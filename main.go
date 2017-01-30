package main;

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"os"
	"encoding/json"
	"io/ioutil"
	"strings"
	"regexp"
	"strconv"
	"time"
)

var users map[string]int;

var rNumber = regexp.MustCompile("-?[0-9]+");
var rMentions = regexp.MustCompile("@[!0-9]+");

func main(){
	bytes, err := ioutil.ReadFile("timezones.json");
	if(err != nil){
		fmt.Println("Note: timezones.json not read: ", err);
		users = make(map[string]int, 0);
	} else {
		err = json.Unmarshal(bytes, &users);
		if(err != nil){
			fmt.Println("Couldn't parse JSON");
			return;
		}
	}

	args := os.Args[1:];
	if(len(args) < 1){
		fmt.Fprintln(os.Stderr, "No token specified!");
		return;
	}
	fmt.Println("Starting...");
	discord, err := discordgo.New("Bot " + args[0]);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "Couldn't start Discord bot");
		return;
	}

	discord.AddHandler(message);
	discord.Open();

	<-make(chan struct{});
}

func save() error{
	bytes, err := json.Marshal(users);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "Could not generate JSON: ", err);
		return err;
	}
	err = ioutil.WriteFile("timezones.json", bytes, 0666);
	if(err != nil){
		fmt.Fprintln(os.Stderr, "Could not save timezones.json: ", err);
		return err;
	}
	return nil;
}

func message(session *discordgo.Session, msg *discordgo.MessageCreate){
	if(msg.Author.Bot){ return; }

	content := strings.ToLower(strings.TrimSpace(msg.Content));
	content = rMentions.ReplaceAllString(content, "");
	if(strings.Contains(content, "my timezone")){
		number := rNumber.FindString(content);
		if(number == ""){
			_, err := session.ChannelMessageSend(msg.ChannelID, "If you meant " +
			"to change your timezone, please supply a number. Otherwise, " +
			"sorry for disturbing!");
			if(err != nil){
				fmt.Fprintln(os.Stderr, "Could not send: ", err);
				return;
			}
			return;
		}

		n, err := strconv.Atoi(number);
		if(err != nil){
			fmt.Fprintln(os.Stderr, "String -> Int error: ", err);
			return;
		}

		users[msg.Author.ID] = n;
		if(save() != nil){
			return;
		}
		_, err = session.ChannelMessageSend(msg.ChannelID, "Set timezone for " +
		msg.Author.Username + " to UTC+" + number);

		if(err != nil){
			fmt.Fprintln(os.Stderr, "Could not send: ", err);
			return;
		}
	} else if(strings.Contains(content, "time")){
		number := rNumber.FindString(content);
		if(number == ""){
			for _, user := range msg.Mentions{
				var err error;
				if timezone, ok := users[user.ID]; ok{
					t := time.Now().UTC().Add(time.Duration(timezone) * time.Hour);
					format := t.Format("03:04:05 PM");

					_, err = session.ChannelMessageSend(msg.ChannelID, "Current " +
					"time for " + user.Username + " is " + format);
				} else {
					_, err = session.ChannelMessageSend(msg.ChannelID, "No " +
					"timezone defined for " + user.Username + "!");
				}

				if(err != nil){
					fmt.Fprintln(os.Stderr, "Could not send: ", err);
					return;
				}
			}
		} else {
			n, err := strconv.Atoi(number);
			if(err != nil){
				fmt.Fprintln(os.Stderr, "String -> Int error: ", err);
				return;
			}

			is_12_h := n <= 12;
			pm := (strings.Contains(content, "pm") || n == 12) && is_12_h;

			n2 := n;
			if(pm && n2 != 12){
				n2 += 12;
			} else if(!pm && n2 == 12){
				n2 = 0;
			}

			if(n2 < 0 || n2 >= 24){
				_, err = session.ChannelMessageSend(msg.ChannelID, "Invalid " +
				"time. Has to be either 0-12am/pm or 0-23.");

				if(err != nil){
					fmt.Fprintln(os.Stderr, "Could not send: ", err);
					return;
				}
				return;
			}

			timezone, ok := users[msg.Author.ID];
			if(!ok){
				_, err = session.ChannelMessageSend(msg.ChannelID, "Your " +
				"timezone isn't set.");

				if(err != nil){
					fmt.Fprintln(os.Stderr, "Could not send: ", err);
					return;
				}
				return;
			}

			for _, user := range msg.Mentions{
				timezone2, ok := users[user.ID];
				if(!ok){
					_, err = session.ChannelMessageSend(msg.ChannelID,
					user.Username + "'s timezone isn't set.");
				} else {
					t := n2 - timezone + timezone2;

					pmstr := "";
					if(pm){
						pmstr = "PM ";
					} else if(is_12_h){
						pmstr = "AM ";
					}

					pmstr2 := "AM";
					if(t >= 12){
						if(t != 12){
							t -= 12;
						}
						pmstr2 = "PM";
					}

					_, err = session.ChannelMessageSend(msg.ChannelID,
					strconv.Itoa(n) + " " + pmstr + "your time is " +
					strconv.Itoa(t) + " " + pmstr2 +
					" for " + user.Username + ".");
				}

				if(err != nil){
					fmt.Fprintln(os.Stderr, "Could not send: ", err);
					return;
				}
			}
		}
	}
}
