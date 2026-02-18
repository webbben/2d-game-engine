package cmd

import (
	"fmt"

	"github.com/webbben/2d-game-engine/data/defs"
	"github.com/webbben/2d-game-engine/dialogv2"
)

const (
	ProfileTest            defs.DialogProfileID = "test1"
	ProfileDefault         defs.DialogProfileID = "default"
	ProfileCharJovis       defs.DialogProfileID = "char_jovis"
	ProfileQ001PrisonGuard defs.DialogProfileID = "Q001_prisonguard"

	TopicRumors defs.TopicID = "rumors"
	TopicJoke   defs.TopicID = "joke"
	TopicTest   defs.TopicID = "lorem_ipsum"
)

func GetDialogProfiles() []defs.DialogProfileDef {
	profiles := []defs.DialogProfileDef{
		{
			ProfileID: ProfileTest,
			Greeting: []defs.DialogResponse{
				{
					Text: "Hello! What might I do for you today?",
				},
			},
			TopicsIDs: []defs.TopicID{TopicRumors, TopicJoke},
		},
		{
			ProfileID: ProfileDefault,
			Greeting: []defs.DialogResponse{
				{
					Text: "(This person has nothing to say, and appears to be lost in thought.)",
				},
			},
		},
		{
			ProfileID: ProfileQ001PrisonGuard,
			Greeting: []defs.DialogResponse{
				{
					ID:   "head_on_up",
					Once: true,
					Text: "Prisoner! This is where you get off. Head on up to the deck and follow the guards orders from there.",
				},
				{
					Text: "Come on now, don't keep them waiting.",
				},
			},
		},
		{
			ProfileID: ProfileCharJovis,
			Greeting: []defs.DialogResponse{
				{
					ID:   "here_comes_the_guard",
					Once: true,
					Text: "Oh, you're awake? I don't think I've seen you so much as stir for the past few days. Tell me, what's your name?",
					NextResponse: &defs.DialogResponse{
						Action: &defs.DialogAction{
							Type:  dialogv2.ActionTypeGetUserInput,
							Scope: dialogv2.ActionScopePlayerName,
							Params: dialogv2.GetUserInputActionParams{
								ModalTitle:        "What's Your Name?",
								ConfirmButtonText: "Confirm",
							},
						},
						Text: fmt.Sprintf("%s, is it? To tell you the truth, I wasn't sure if you'd make it. You've been lying there motionless for days, or at least ever since they threw me in here.", dialogv2.VarPlayerName),
						NextResponse: &defs.DialogResponse{
							Text: "Say, how did you wind up here? A common criminal? Or perhaps an enemy soldier, captured in the chaos of battle?",
							NextResponse: &defs.DialogResponse{
								Text: "...",
								NextResponse: &defs.DialogResponse{
									Text: "You're not sure, you say? Well, You must've been dealt one serious blow to the head on a battlefield. That would explain your funny accent, too.",
									NextResponse: &defs.DialogResponse{
										Text: `
										I've heard we've been sailing through the Adriatic for a day or so now, and I have a feeling we'll be reaching our destination soon.
										Where to? I'm not sure. But I do know that nothing great is waiting for us on land. I imagine we'll be sold into slavery.
										We must not be headed to the plantations though, because surely they would've sailed for Sicily, or maybe Africa...
										`,
										NextResponse: &defs.DialogResponse{
											Text: "Gods, I hope we aren't headed for the mines. That is a truly terrible fate, my friend. Between you and me, maybe we ought to -",
											NextResponse: &defs.DialogResponse{
												Text: fmt.Sprintf(" - Oh, Quiet! Here comes the guard! I wish you the best of luck, %s.", dialogv2.VarPlayerName),
											},
											Effects: []defs.DialogEffect{
												dialogv2.EventEffect{
													Event: defs.Event{
														Type: Q001GuardApproaching,
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				{
					Text: "Not now - you better do what they tell you.",
				},
			},
		},
	}

	return profiles
}

func GetDialogTopics() []defs.DialogTopic {
	topics := []defs.DialogTopic{
		{
			ID:     TopicRumors,
			Prompt: "Rumors",
			Responses: []defs.DialogResponse{
				{
					Text: "If you go to the forum after midnight, some say you find some shadowy figures hanging about. I'd steer clear away from there if I were you.",
				},
			},
		},
		{
			ID:     TopicJoke,
			Prompt: "Joke",
			Responses: []defs.DialogResponse{
				{
					Text: "Why did the chicken cross the road?",
					Replies: []defs.DialogReply{
						{
							Text: "To get to the other side?",
							NextResponse: &defs.DialogResponse{
								Text:       "No dummy! He was running away from the town butcher!",
								NextTopics: []defs.TopicID{TopicTest},
							},
						},
						{
							Text: "I don't know, why?",
							NextResponse: &defs.DialogResponse{
								Text: "Come on, not even a guess?",
							},
						},
					},
				},
			},
		},
		{
			ID:     TopicTest,
			Prompt: "Lorem Ipsum",
			Responses: []defs.DialogResponse{
				{
					Text: `
Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum suscipit urna ex, laoreet gravida risus blandit consectetur. Nullam pulvinar, enim et commodo fringilla, nulla magna tempor enim, quis elementum est mauris at mauris. Nam non ligula a enim sollicitudin luctus. Sed aliquet maximus erat aliquam iaculis. In tempus sapien nisi. Etiam tortor massa, tristique nec ex in, imperdiet dignissim nisi. Vivamus id mi at dolor suscipit luctus. In nec lacus et elit rhoncus cursus. Sed porttitor, dui eu ornare fringilla, dui risus placerat eros, nec porta sem justo sit amet neque. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Cras vel congue tortor. Mauris aliquet molestie massa, venenatis volutpat justo convallis vestibulum.
Integer nisi ligula, volutpat feugiat eros ut, cursus eleifend est. Quisque gravida sit amet dui vitae pellentesque. Morbi interdum facilisis tellus aliquam egestas. Nunc posuere nunc neque, a sagittis elit ultricies eget. Aliquam vel dignissim dui. Quisque mollis massa nibh, id dignissim ante semper eu. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Vivamus sed accumsan felis.
Sed velit neque, eleifend quis arcu sed, fringilla varius augue. Aenean interdum ornare consectetur. Mauris eleifend mauris erat, non luctus nunc venenatis at. Aliquam ultricies dolor sed odio iaculis, id gravida ipsum faucibus. Morbi vitae rutrum nisl. Praesent lorem leo, tincidunt eu felis quis, ornare blandit nisl. Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Nulla convallis diam sed elementum posuere. Morbi vulputate urna vitae quam gravida pellentesque. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. Maecenas a tincidunt turpis, sed luctus enim. Sed venenatis quam non velit viverra ultricies. Maecenas tristique lacinia mauris id interdum. Suspendisse tempus enim arcu, id fringilla mauris consectetur quis. Vestibulum placerat ante lacus, sed tempor mi pharetra ut. Aenean metus augue, vestibulum in tincidunt sed, tempus nec lacus.
Quisque sollicitudin auctor magna, a pharetra justo consequat sed. Nulla mi leo, ultricies in neque ac, elementum posuere neque. Praesent et maximus est. Etiam pulvinar velit a felis bibendum molestie. Donec faucibus mi in elit dapibus fermentum. Quisque vestibulum libero quis lacus tincidunt volutpat. Nullam posuere mauris odio, vitae venenatis tortor sodales ac. Donec porta massa eu vehicula dapibus. Phasellus vulputate placerat urna, nec feugiat est porta sed. Curabitur ac turpis sem. Morbi nisi turpis, dignissim eu nisi at, mollis posuere lorem. Maecenas pretium congue lectus, ut tempus dolor ornare vitae.
In et aliquet orci. Curabitur pharetra sit amet felis et faucibus. Morbi vitae massa quam. Aliquam porta, nulla quis egestas lobortis, magna diam ultrices felis, a scelerisque justo diam a est. Vestibulum nisi leo, placerat ut laoreet vel, iaculis id sapien. Ut et placerat lectus. Aliquam erat volutpat. Fusce finibus sapien quis justo lobortis feugiat.
			`,
				},
			},
		},
	}

	return topics
}
