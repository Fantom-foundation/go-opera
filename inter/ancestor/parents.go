package ancestor

import (
	"math/rand"
	"sort"
	"time"

	"github.com/Fantom-foundation/go-lachesis/hash"
	"github.com/Fantom-foundation/go-lachesis/inter/idx"
	"github.com/Fantom-foundation/go-lachesis/inter/pos"
	"github.com/Fantom-foundation/go-lachesis/vector"
)

// SearchStrategy defines a criteria used to estimate the "best" subset of parents to emit event with.
type SearchStrategy interface {
	// Init must be called before using the strategy
	Init(selfParent *hash.Event)
	// Find chooses the hash from the specified options
	Find(options hash.Events) hash.Event
}

// FindBestParents returns estimated parents subset, according to provided strategy
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

// CasualityStrategy uses vector clock to check which parents observe "more" than others
// The strategy uses "observing more" as a search criteria
type CasualityStrategy struct {
	vecClock   *vector.Index
	template   vector.HighestBeforeSeq
	validators *pos.Validators
}

// NewCasualityStrategy creates new CasualityStrategy with provided vector clock
func NewCasualityStrategy(vecClock *vector.Index, validators *pos.Validators) *CasualityStrategy {
	return &CasualityStrategy{
		vecClock:   vecClock,
		validators: validators,
	}
}

type eventScore struct {
	event hash.Event
	score idx.Event
	vec   vector.HighestBeforeSeq
}

// Init must be called before using the strategy
func (st *CasualityStrategy) Init(selfParent *hash.Event) {
	if selfParent != nil {
		// we start searching by comparing with self-parent
		st.template = st.vecClock.GetHighestBeforeAllBranches(*selfParent)
	}
}

// Find chooses the hash from the specified options
func (st *CasualityStrategy) Find(options hash.Events) hash.Event {
	scores := make([]eventScore, 0, 100)

	// estimate score of each option as number of validators it observes higher than provided template
	for _, id := range options {
		score := eventScore{}
		score.event = id
		score.vec = st.vecClock.GetHighestBeforeAllBranches(id)
		if st.template == nil {
			st.template = vector.NewHighestBeforeSeq(st.validators.Len()) // nothing observes
		}
		for creatorIdx := idx.Validator(0); creatorIdx < idx.Validator(st.validators.Len()); creatorIdx++ {
			my := st.template.Get(creatorIdx)
			his := score.vec.Get(creatorIdx)

			// observes higher
			if his.Seq > my.Seq && !my.IsForkDetected() {
				score.score++
			}
			// observes a fork
			if his.IsForkDetected() && !my.IsForkDetected() {
				score.score++
			}
		}
		scores = append(scores, score)
	}

	// take the option with best score
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

// RandomStrategy is used in tests, when vector clock isn't available
type RandomStrategy struct {
	r *rand.Rand
}

func NewRandomStrategy(r *rand.Rand) *RandomStrategy {
	if r == nil {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	}
	return &RandomStrategy{
		r: r,
	}
}

func (st *RandomStrategy) Init(myLast *hash.Event) {}

// Find chooses the hash from the specified options
func (st *RandomStrategy) Find(heads hash.Events) hash.Event {
	return heads[st.r.Intn(len(heads))]
}
