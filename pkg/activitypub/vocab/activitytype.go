/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package vocab

import (
	"net/url"
)

// ActivityType defines an 'activity'.
type ActivityType struct {
	*ObjectType

	activity *activityType
}

type activityType struct {
	Actor  *URLProperty    `json:"actor,omitempty"`
	Target *ObjectProperty `json:"target,omitempty"`
	Object *ObjectProperty `json:"object,omitempty"`
	Result *ObjectProperty `json:"result,omitempty"`
}

// Actor returns the actor for the activity.
func (t *ActivityType) Actor() *url.URL {
	if t.activity.Actor == nil {
		return nil
	}

	return t.activity.Actor.URL()
}

// Target returns the target of the activity.
func (t *ActivityType) Target() *ObjectProperty {
	return t.activity.Target
}

// Object returns the object of the activity.
func (t *ActivityType) Object() *ObjectProperty {
	return t.activity.Object
}

// Result returns the result.
func (t *ActivityType) Result() *ObjectProperty {
	return t.activity.Result
}

// MarshalJSON marshals the activity.
func (t *ActivityType) MarshalJSON() ([]byte, error) {
	return MarshalJSON(t.ObjectType, t.activity)
}

// UnmarshalJSON unmarshals the activity.
func (t *ActivityType) UnmarshalJSON(bytes []byte) error {
	t.ObjectType = NewObject()
	t.activity = &activityType{}

	return UnmarshalJSON(bytes, t.ObjectType, t.activity)
}

// NewCreateActivity returns a new 'Create' activity.
func NewCreateActivity(id string, obj *ObjectProperty, opts ...Opt) *ActivityType {
	options := NewOptions(opts...)

	return &ActivityType{
		ObjectType: NewObject(
			WithContext(getContexts(options, ContextActivityStreams)...),
			WithID(id),
			WithType(TypeCreate),
			WithTo(options.To...),
			WithPublishedTime(options.Published),
		),
		activity: &activityType{
			Actor:  NewURLProperty(options.Actor),
			Target: options.Target,
			Object: obj,
		},
	}
}

// NewAnnounceActivity returns a new 'Announce' activity.
func NewAnnounceActivity(id string, obj *ObjectProperty, opts ...Opt) *ActivityType {
	options := NewOptions(opts...)

	return &ActivityType{
		ObjectType: NewObject(
			WithContext(getContexts(options, ContextActivityStreams)...),
			WithID(id),
			WithType(TypeAnnounce),
			WithTo(options.To...),
			WithPublishedTime(options.Published),
		),
		activity: &activityType{
			Actor:  NewURLProperty(options.Actor),
			Object: obj,
		},
	}
}

// NewFollowActivity returns a new 'Follow' activity.
func NewFollowActivity(id string, obj *ObjectProperty, opts ...Opt) *ActivityType {
	options := NewOptions(opts...)

	return &ActivityType{
		ObjectType: NewObject(
			WithContext(getContexts(options, ContextActivityStreams)...),
			WithID(id),
			WithType(TypeFollow),
			WithTo(options.To...),
		),
		activity: &activityType{
			Actor:  NewURLProperty(options.Actor),
			Object: obj,
		},
	}
}

// NewAcceptActivity returns a new 'Accept' activity.
func NewAcceptActivity(id string, obj *ObjectProperty, opts ...Opt) *ActivityType {
	options := NewOptions(opts...)

	return &ActivityType{
		ObjectType: NewObject(
			WithContext(getContexts(options, ContextActivityStreams)...),
			WithID(id),
			WithType(TypeAccept),
			WithTo(options.To...),
		),
		activity: &activityType{
			Actor:  NewURLProperty(options.Actor),
			Object: obj,
		},
	}
}

// NewRejectActivity returns a new 'Reject' activity.
func NewRejectActivity(id string, obj *ObjectProperty, opts ...Opt) *ActivityType {
	options := NewOptions(opts...)

	return &ActivityType{
		ObjectType: NewObject(
			WithContext(getContexts(options, ContextActivityStreams)...),
			WithID(id),
			WithType(TypeReject),
			WithTo(options.To...),
		),
		activity: &activityType{
			Actor:  NewURLProperty(options.Actor),
			Object: obj,
		},
	}
}

// NewLikeActivity returns a new 'Like' activity.
func NewLikeActivity(id string, obj *ObjectProperty, opts ...Opt) *ActivityType {
	options := NewOptions(opts...)

	return &ActivityType{
		ObjectType: NewObject(
			WithContext(getContexts(options, ContextActivityStreams)...),
			WithID(id),
			WithType(TypeLike),
			WithTo(options.To...),
			WithStartTime(options.StartTime),
			WithEndTime(options.EndTime),
		),
		activity: &activityType{
			Actor:  NewURLProperty(options.Actor),
			Object: obj,
			Result: options.Result,
		},
	}
}

// NewOfferActivity returns a new 'Offer' activity.
func NewOfferActivity(id string, obj *ObjectProperty, opts ...Opt) *ActivityType {
	options := NewOptions(opts...)

	return &ActivityType{
		ObjectType: NewObject(
			WithContext(getContexts(options, ContextActivityStreams)...),
			WithID(id),
			WithType(TypeOffer),
			WithTo(options.To...),
			WithStartTime(options.StartTime),
			WithEndTime(options.EndTime),
		),
		activity: &activityType{
			Actor:  NewURLProperty(options.Actor),
			Object: obj,
		},
	}
}
