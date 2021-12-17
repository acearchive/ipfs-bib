package pattern_test

import (
	"github.com/frawleyskid/ipfs-bib/config/pattern"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Parser", func() {
	var parser pattern.Parser

	BeforeEach(func() {
		parser = pattern.NewParser(map[pattern.Var]string{
			'a': "apple",
			'b': "banana",
		})
	})

	Describe("Parse", func() {
		It("parses a pattern with a variable in the middle", func() {
			result, err := parser.Parse("before %a after")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("before apple after"))
		})

		It("parses a pattern with a variable at the beginning", func() {
			result, err := parser.Parse("%a after")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("apple after"))
		})

		It("parses a pattern with a variable at the end", func() {
			result, err := parser.Parse("before %a")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("before apple"))
		})

		It("detects literal % symbols", func() {
			result, err := parser.Parse("before %% after")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("before % after"))
		})

		It("parses a pattern with multiple variables", func() {
			result, err := parser.Parse("before %a middle %b after")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("before apple middle banana after"))
		})

		It("parses a pattern with variables at the beginning and middle", func() {
			result, err := parser.Parse("%a middle %b after")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("apple middle banana after"))
		})

		It("parses a pattern with variables at the middle and end", func() {
			result, err := parser.Parse("before %a middle %b")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("before apple middle banana"))
		})

		It("parses a pattern with multiple adjacent variables", func() {
			result, err := parser.Parse("before %a%b after")
			Expect(err).ToNot(HaveOccurred())
			Expect(result).To(Equal("before applebanana after"))
		})

		It("returns err when a variable doesn't exist", func() {
			_, err := parser.Parse("before %c after")
			Expect(err).To(MatchError(pattern.ErrInvalidPatternVar))
		})
	})
})
