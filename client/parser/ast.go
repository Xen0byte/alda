package parser

import (
	"fmt"

	text "alda.io/client/text"

	model "alda.io/client/model"
)

// An ASTNodeType is a type of AST node output by the parser.
type ASTNodeType int

const (
	AtMarkerNode ASTNodeType = iota
	BarlineNode
	ChordNode
	CramNode
	DurationNode
	EventSequenceNode
	FlatNode
	ImplicitPartNode
	LispListNode
	LispNumberNode
	LispQuotedFormNode
	LispStringNode
	LispSymbolNode
	MarkerNode
	NaturalNode
	NoteAccidentalsNode
	NoteLengthMsNode
	NoteLengthNode
	NoteLetterAndAccidentalsNode
	NoteLetterNode
	NoteNode
	OctaveDownNode
	OctaveSetNode
	OctaveUpNode
	OnRepetitionsNode
	PartAliasNode
	PartDeclarationNode
	PartNameNode
	PartNamesNode
	PartNode
	RepeatNode
	RepetitionsNode
	RestNode
	RootNode
	SharpNode
	TieNode
	TimesNode
	VariableDefinitionNode
	VariableNameNode
	VariableReferenceNode
	VoiceNode
	VoiceGroupEndMarkerNode
	VoiceGroupNode
	VoiceNumberNode
)

type ASTNode struct {
	Type          ASTNodeType
	Literal       interface{}
	Children      []ASTNode
	SourceContext model.AldaSourceContext
}

func (nt ASTNodeType) String() string {
	switch nt {
	case AtMarkerNode:
		return "AtMarkerNode"
	case BarlineNode:
		return "BarlineNode"
	case ChordNode:
		return "ChordNode"
	case CramNode:
		return "CramNode"
	case DurationNode:
		return "DurationNode"
	case EventSequenceNode:
		return "EventSequenceNode"
	case FlatNode:
		return "FlatNode"
	case ImplicitPartNode:
		return "ImplicitPartNode"
	case LispListNode:
		return "LispListNode"
	case LispNumberNode:
		return "LispNumberNode"
	case LispQuotedFormNode:
		return "LispQuotedFormNode"
	case LispStringNode:
		return "LispStringNode"
	case LispSymbolNode:
		return "LispSymbolNode"
	case MarkerNode:
		return "MarkerNode"
	case NaturalNode:
		return "NaturalNode"
	case NoteAccidentalsNode:
		return "NoteAccidentalsNode"
	case NoteLengthMsNode:
		return "NoteLengthMsNode"
	case NoteLengthNode:
		return "NoteLengthNode"
	case NoteLetterAndAccidentalsNode:
		return "NoteLetterAndAccidentalsNode"
	case NoteLetterNode:
		return "NoteLetterNode"
	case NoteNode:
		return "NoteNode"
	case OctaveDownNode:
		return "OctaveDownNode"
	case OctaveSetNode:
		return "OctaveSetNode"
	case OctaveUpNode:
		return "OctaveUpNode"
	case OnRepetitionsNode:
		return "OnRepetitionsNode"
	case PartAliasNode:
		return "PartAliasNode"
	case PartDeclarationNode:
		return "PartDeclarationNode"
	case PartNameNode:
		return "PartNameNode"
	case PartNamesNode:
		return "PartNamesNode"
	case PartNode:
		return "PartNode"
	case RepeatNode:
		return "RepeatNode"
	case RepetitionsNode:
		return "RepetitionsNode"
	case RestNode:
		return "RestNode"
	case RootNode:
		return "RootNode"
	case SharpNode:
		return "SharpNode"
	case TieNode:
		return "TieNode"
	case TimesNode:
		return "TimesNode"
	case VariableDefinitionNode:
		return "VariableDefinitionNode"
	case VariableNameNode:
		return "VariableNameNode"
	case VariableReferenceNode:
		return "VariableReferenceNode"
	case VoiceNode:
		return "VoiceNode"
	case VoiceGroupEndMarkerNode:
		return "VoiceGroupEndMarkerNode"
	case VoiceGroupNode:
		return "VoiceGroupNode"
	case VoiceNumberNode:
		return "VoiceNumberNode"
	default:
		return fmt.Sprintf("%d (String not implemented)", nt)
	}
}

// PrettyPrint prints an ASTNode in a concise, easy to visualize format.
func (node ASTNode) PrettyPrint() {
	var recursivelyPrettyPrint func(ASTNode, int)
	recursivelyPrettyPrint = func(node ASTNode, indentLevel int) {
		maybeLiteral := ""
		if node.Literal != nil {
			literal := node.Literal
			if node.Type == NoteLetterNode {
				literal = string(node.Literal.(rune))
			}

			maybeLiteral = fmt.Sprintf(": %#v", literal)
		}

		fmt.Println(
			text.Indent(indentLevel, node.Type.String()+maybeLiteral),
		)

		for _, child := range node.Children {
			recursivelyPrettyPrint(child, indentLevel+1)
		}
	}

	recursivelyPrettyPrint(node, 0)
}

func (node ASTNode) expectChildren() error {
	if len(node.Children) == 0 {
		return fmt.Errorf("%s has no children", node.Type.String())
	}

	return nil
}

func (node ASTNode) expectNChildren(expectedChildren int) error {
	actualChildren := len(node.Children)

	if actualChildren != expectedChildren {
		return fmt.Errorf(
			"expected %s to have %d children, but it has %d",
			node.Type.String(),
			expectedChildren,
			actualChildren,
		)
	}

	return nil
}

func (node ASTNode) expectNodeType(expectedType ASTNodeType) (ASTNode, error) {
	if node.Type != expectedType {
		return ASTNode{}, fmt.Errorf(
			"expected %s node, but got %s node",
			expectedType.String(),
			node.Type.String(),
		)
	}

	return node, nil
}

func errUnexpectedNodeChild(parentType ASTNodeType, childType ASTNodeType) error {
	return fmt.Errorf(
		"unexpected %s child: %s",
		parentType.String(), childType.String(),
	)
}

func duration(node ASTNode) (model.Duration, error) {
	duration := model.Duration{}

	for _, componentNode := range node.Children {
		switch componentNode.Type {
		default:
			return model.Duration{}, errUnexpectedNodeChild(node.Type, componentNode.Type)
		case BarlineNode:
			barline := model.Barline{SourceContext: componentNode.SourceContext}
			duration.Components = append(duration.Components, barline)
		case NoteLengthNode:
			literal := componentNode.Literal.(noteLength)
			noteLength := model.NoteLength{
				Denominator: literal.denominator,
				Dots:        literal.dots,
			}
			duration.Components = append(duration.Components, noteLength)
		case NoteLengthMsNode:
			literal := componentNode.Literal.(float64)
			noteLengthMs := model.NoteLengthMs{Quantity: literal}
			duration.Components = append(duration.Components, noteLengthMs)
		}
	}

	return duration, nil
}

func (node ASTNode) Updates() ([]model.ScoreUpdate, error) {
	concatChildUpdates := func(node ASTNode) ([]model.ScoreUpdate, error) {
		updates := []model.ScoreUpdate{}

		for _, child := range node.Children {
			childUpdates, err := child.Updates()
			if err != nil {
				return nil, err
			}

			updates = append(updates, childUpdates...)
		}

		return updates, nil
	}

	switch node.Type {

	case AtMarkerNode:
		return []model.ScoreUpdate{
			model.AtMarker{
				SourceContext: node.SourceContext,
				Name:          node.Literal.(string),
			},
		}, nil

	case BarlineNode:
		return []model.ScoreUpdate{
			model.Barline{SourceContext: node.SourceContext},
		}, nil

	case ChordNode:
		if err := node.expectChildren(); err != nil {
			return nil, err
		}

		updates, err := concatChildUpdates(node)
		if err != nil {
			return nil, err
		}

		return []model.ScoreUpdate{
			model.Chord{
				SourceContext: node.SourceContext,
				Events:        updates,
			},
		}, nil

	case CramNode:
		if err := node.expectChildren(); err != nil {
			return nil, err
		}

		eventsNode, err := node.Children[0].expectNodeType(EventSequenceNode)
		if err != nil {
			return nil, err
		}

		events, err := concatChildUpdates(eventsNode)
		if err != nil {
			return nil, err
		}

		cram := model.Cram{
			SourceContext: node.SourceContext,
			Events:        events,
		}

		if len(node.Children) > 1 {
			durationNode, err := node.Children[1].expectNodeType(DurationNode)
			if err != nil {
				return nil, err
			}

			dur, err := duration(durationNode)
			if err != nil {
				return nil, err
			}
			cram.Duration = dur
		}

		return []model.ScoreUpdate{cram}, nil

	case EventSequenceNode:
		updates, err := concatChildUpdates(node)
		if err != nil {
			return nil, err
		}

		return []model.ScoreUpdate{
			model.EventSequence{
				SourceContext: node.SourceContext,
				Events:        updates,
			},
		}, nil

	case ImplicitPartNode:
		if err := node.expectNChildren(1); err != nil {
			return nil, err
		}

		events, err := node.Children[0].expectNodeType(EventSequenceNode)
		if err != nil {
			return nil, err
		}

		return concatChildUpdates(events)

	case LispListNode:
		var lispForm func(ASTNode) (model.LispForm, error)
		lispForm = func(node ASTNode) (model.LispForm, error) {
			switch node.Type {
			case LispListNode:
				list := model.LispList{SourceContext: node.SourceContext}

				for _, child := range node.Children {
					form, err := lispForm(child)
					if err != nil {
						return nil, err
					}

					list.Elements = append(list.Elements, form)
				}

				return list, nil

			case LispNumberNode:
				return model.LispNumber{
					SourceContext: node.SourceContext,
					Value:         node.Literal.(float64),
				}, nil

			case LispQuotedFormNode:
				if err := node.expectNChildren(1); err != nil {
					return nil, err
				}

				form, err := lispForm(node.Children[0])
				if err != nil {
					return nil, err
				}

				return model.LispQuotedForm{
					SourceContext: node.SourceContext,
					Form:          form,
				}, nil

			case LispStringNode:
				return model.LispString{
					SourceContext: node.SourceContext,
					Value:         node.Literal.(string),
				}, nil

			case LispSymbolNode:
				return model.LispSymbol{
					SourceContext: node.SourceContext,
					Name:          node.Literal.(string),
				}, nil
			}

			return nil, fmt.Errorf(
				"unexpected %s node inside of Lisp form", node.Type.String(),
			)
		}

		list, err := lispForm(node)
		if err != nil {
			return nil, err
		}

		return []model.ScoreUpdate{list.(model.LispList)}, nil

	case MarkerNode:
		return []model.ScoreUpdate{
			model.Marker{
				SourceContext: node.SourceContext,
				Name:          node.Literal.(string),
			},
		}, nil

	case NoteNode:
		if err := node.expectChildren(); err != nil {
			return nil, err
		}

		laaNode, err := node.Children[0].expectNodeType(NoteLetterAndAccidentalsNode)
		if err != nil {
			return nil, err
		}

		letterNode, err := laaNode.Children[0].expectNodeType(NoteLetterNode)
		if err != nil {
			return nil, err
		}

		noteLetter, err := model.NewNoteLetter(letterNode.Literal.(rune))
		if err != nil {
			return nil, err
		}

		laa := model.LetterAndAccidentals{NoteLetter: noteLetter}

		if len(laaNode.Children) > 1 {
			accidentalsNode, err := laaNode.Children[1].expectNodeType(
				NoteAccidentalsNode,
			)
			if err != nil {
				return nil, err
			}

			for _, child := range accidentalsNode.Children {
				switch child.Type {
				default:
					return nil, errUnexpectedNodeChild(accidentalsNode.Type, child.Type)
				case FlatNode:
					laa.Accidentals = append(laa.Accidentals, model.Flat)
				case NaturalNode:
					laa.Accidentals = append(laa.Accidentals, model.Natural)
				case SharpNode:
					laa.Accidentals = append(laa.Accidentals, model.Sharp)
				}
			}
		}

		note := model.Note{
			SourceContext: node.SourceContext,
			Pitch:         laa,
		}

		if len(node.Children) > 1 {
			for _, child := range node.Children[1:] {
				switch child.Type {
				default:
					return nil, errUnexpectedNodeChild(node.Type, child.Type)
				case DurationNode:
					dur, err := duration(child)
					if err != nil {
						return nil, err
					}
					note.Duration = dur
				case TieNode:
					note.Slurred = true
				}
			}
		}

		return []model.ScoreUpdate{note}, nil

	case OctaveDownNode:
		return []model.ScoreUpdate{
			model.AttributeUpdate{
				SourceContext: node.SourceContext,
				PartUpdate:    model.OctaveDown{},
			},
		}, nil

	case OctaveSetNode:
		return []model.ScoreUpdate{
			model.AttributeUpdate{
				SourceContext: node.SourceContext,
				PartUpdate:    model.OctaveSet{OctaveNumber: node.Literal.(int32)},
			},
		}, nil

	case OctaveUpNode:
		return []model.ScoreUpdate{
			model.AttributeUpdate{
				SourceContext: node.SourceContext,
				PartUpdate:    model.OctaveUp{},
			},
		}, nil

	case OnRepetitionsNode:
		if err := node.expectNChildren(2); err != nil {
			return nil, err
		}

		eventNode := node.Children[0]

		eventUpdates, err := eventNode.Updates()
		if err != nil {
			return nil, err
		}

		var event model.ScoreUpdate

		// I don't _think_ `eventUpdates` will ever contain more than one update,
		// but just in case...
		if len(eventUpdates) == 1 {
			event = eventUpdates[0]
		} else {
			event = model.EventSequence{
				SourceContext: eventNode.SourceContext,
				Events:        eventUpdates,
			}
		}

		repetitions, err := node.Children[1].expectNodeType(RepetitionsNode)
		if err != nil {
			return nil, err
		}

		return []model.ScoreUpdate{
			model.OnRepetitions{
				SourceContext: node.SourceContext,
				Event:         event,
				Repetitions:   repetitions.Literal.([]model.RepetitionRange),
			},
		}, nil

	case PartNode:
		if err := node.expectChildren(); err != nil {
			return nil, err
		}

		partDeclNode, err := node.Children[0].expectNodeType(PartDeclarationNode)
		if err != nil {
			return nil, err
		}

		partDecl := model.PartDeclaration{SourceContext: node.SourceContext}

		partNames, err := partDeclNode.Children[0].expectNodeType(PartNamesNode)
		if err != nil {
			return nil, err
		}

		if err := partNames.expectChildren(); err != nil {
			return nil, err
		}

		for _, child := range partNames.Children {
			partName, err := child.expectNodeType(PartNameNode)
			if err != nil {
				return nil, err
			}

			partDecl.Names = append(partDecl.Names, partName.Literal.(string))
		}

		if len(partDeclNode.Children) > 1 {
			partAlias, err := partDeclNode.Children[1].expectNodeType(PartAliasNode)
			if err != nil {
				return nil, err
			}

			partDecl.Alias = partAlias.Literal.(string)
		}

		updates := []model.ScoreUpdate{partDecl}

		if len(node.Children) > 1 {
			events, err := node.Children[1].expectNodeType(EventSequenceNode)
			if err != nil {
				return nil, err
			}

			eventUpdates, err := concatChildUpdates(events)
			if err != nil {
				return nil, err
			}

			updates = append(updates, eventUpdates...)
		}

		return updates, nil

	case RepeatNode:
		if err := node.expectNChildren(2); err != nil {
			return nil, err
		}

		eventNode := node.Children[0]

		times, err := node.Children[1].expectNodeType(TimesNode)
		if err != nil {
			return nil, err
		}

		eventUpdates, err := eventNode.Updates()
		if err != nil {
			return nil, err
		}

		var event model.ScoreUpdate

		// I don't _think_ `eventUpdates` will ever contain more than one update,
		// but just in case...
		if len(eventUpdates) == 1 {
			event = eventUpdates[0]
		} else {
			event = model.EventSequence{
				SourceContext: eventNode.SourceContext,
				Events:        eventUpdates,
			}
		}

		return []model.ScoreUpdate{
			model.Repeat{
				SourceContext: eventNode.SourceContext,
				Event:         event,
				Times:         times.Literal.(int32),
			},
		}, nil

	case RestNode:
		rest := model.Rest{SourceContext: node.SourceContext}

		switch len(node.Children) {
		case 0:
			break
		case 1:
			durationNode, err := node.Children[0].expectNodeType(DurationNode)
			if err != nil {
				return nil, err
			}

			dur, err := duration(durationNode)
			if err != nil {
				return nil, err
			}
			rest.Duration = dur
		default:
			return nil, fmt.Errorf("rest node must have either 0 or 1 children")
		}

		return []model.ScoreUpdate{rest}, nil

	case RootNode:
		return concatChildUpdates(node)

	case VariableDefinitionNode:
		if err := node.expectNChildren(2); err != nil {
			return nil, err
		}

		variableName, err := node.Children[0].expectNodeType(VariableNameNode)
		if err != nil {
			return nil, err
		}

		definition := model.VariableDefinition{
			SourceContext: node.SourceContext,
			VariableName:  variableName.Literal.(string),
		}

		eventsNode, err := node.Children[1].expectNodeType(EventSequenceNode)
		if err != nil {
			return nil, err
		}

		events, err := concatChildUpdates(eventsNode)
		if err != nil {
			return nil, err
		}

		definition.Events = events

		return []model.ScoreUpdate{definition}, nil

	case VariableReferenceNode:
		return []model.ScoreUpdate{
			model.VariableReference{
				SourceContext: node.SourceContext,
				VariableName:  node.Literal.(string),
			},
		}, nil

	case VoiceGroupNode:
		return concatChildUpdates(node)

	case VoiceGroupEndMarkerNode:
		return []model.ScoreUpdate{
			model.VoiceGroupEndMarker{
				SourceContext: node.SourceContext,
			},
		}, nil

	case VoiceNode:
		if err := node.expectNChildren(2); err != nil {
			return nil, err
		}

		voiceNumber, err := node.Children[0].expectNodeType(VoiceNumberNode)
		if err != nil {
			return nil, err
		}

		voiceEvents, err := node.Children[1].expectNodeType(EventSequenceNode)
		if err != nil {
			return nil, err
		}

		voiceEventUpdates, err := concatChildUpdates(voiceEvents)
		if err != nil {
			return nil, err
		}

		return append(
			[]model.ScoreUpdate{
				model.VoiceMarker{
					SourceContext: voiceNumber.SourceContext,
					VoiceNumber:   voiceNumber.Literal.(int32),
				},
			},
			voiceEventUpdates...,
		), nil

	case VoiceNumberNode:
		return []model.ScoreUpdate{
			model.VoiceMarker{
				SourceContext: node.SourceContext,
				VoiceNumber:   node.Literal.(int32),
			},
		}, nil
	}

	return nil, fmt.Errorf(
		"Updates() not implemented for AST node type '%s'",
		node.Type,
	)
}