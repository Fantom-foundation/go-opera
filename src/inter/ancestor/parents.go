package ancestor

import (
	"math/rand"
	"sort"

	"github.com/Fantom-foundation/go-lachesis/src/hash"
	"github.com/Fantom-foundation/go-lachesis/src/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/src/vector"
)

type SearchStrategy interface {
	Init(selfParent *hash.Event)
	Find(heads hash.Events) hash.Event
}

// max is max num of parents to link with (including self-parent)
// returns set of parents to link, len(res) <= max
func FindBestParents(max int, options hash.Events, selfParent *hash.Event, strategy SearchStrategy) (*hash.Event, hash.Events) {
	optionsSet := options.Set()
	parents := make(hash.Events, 0, max)
	if selfParent != nil {
		parents = append(parents, *selfParent)
		optionsSet.Erase(*selfParent)
	}

	strategy.Init(selfParent)

	for len(parents) < max && len(optionsSet) > 0 {
		best := strategy.Find(optionsSet.Slice())
		parents = append(parents, best)
		optionsSet.Erase(best)
	}

	return selfParent, parents
}

/*
 * CasualityStrategy
 */

type CasualityStrategy struct {
	vecClock *vector.Index
	template vector.HighestBeforeSeq
}

func NewCasualityStrategy(vecClock *vector.Index) *CasualityStrategy {
	return &CasualityStrategy{
		vecClock: vecClock,
	}
}

type eventScore struct {
	event hash.Event
	score idx.Event
	vec   vector.HighestBeforeSeq
}

func (st *CasualityStrategy) Init(selfParent *hash.Event) {
	if selfParent != nil {
		// we start searching by comparing with self-parent
		st.template = st.vecClock.GetHighestBeforeSeq(*selfParent)
	}
}

func (st *CasualityStrategy) Find(options hash.Events) hash.Event {
	scores := make([]eventScore, 0, 100)

	// estimate score of each option as number of members it causes higher than provided template
	for _, id := range options {
		score := eventScore{}
		score.event = id
		score.vec = st.vecClock.GetHighestBeforeSeq(id)
		if st.template == nil {
			st.template = vector.NewHighestBeforeSeq(int(score.vec.MembersNum())) // nothing causes
		}
		for n := idx.Member(0); n < score.vec.MembersNum(); n++ {
			my := st.template.Get(n)
			his := score.vec.Get(n)

			// causes higher
			if his.Seq > my.Seq && !my.IsForkDetected {
				score.score += 1
			}
			// causes a fork
			if his.IsForkDetected && !my.IsForkDetected {
				score.score += 1
			}
		}
		scores = append(scores, score)
	}

	// take the head with best score
	sort.Slice(scores, func(i, j int) bool {
		a, b := scores[i], scores[j]
		return a.score < b.score
	})
	// memorize its template for next calls
	st.template = scores[0].vec
	return scores[0].event
}

/*
 * RandomStrategy
 */

type RandomStrategy struct {
	r *rand.Rand
}

func NewRandomStrategy(r *rand.Rand) *RandomStrategy {
	if r == nil {
		r = rand.New(rand.NewSource(0))
	}
	return &RandomStrategy{
		r: r,
	}
}

func (st *RandomStrategy) Init(myLast *hash.Event) {}

func (st *RandomStrategy) Find(heads hash.Events) hash.Event {
	return heads[st.r.Intn(len(heads))]
}
