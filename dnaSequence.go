package bioutils

import (
	"errors"
	"reflect"
	"strings"
)

// Represents a sequence of nucleotides
type dnaSequence []nucleotide
type sequenceGroup []dnaSequence

// Takes in a string representation of the nucleotide bases and returns a
// completely checked dna sequence
func CreateDNASequence(uncheckedSequence string) (dnaSequence, error) {
	newSequence := []nucleotide{}
	for _, currentBaseLetter := range uncheckedSequence {
		curNucleotide := nucleotide(currentBaseLetter)
		if !curNucleotide.ValidNucleotide() {
			return newSequence, errors.New("Invalid base " + string(currentBaseLetter))
		}
		newSequence = append(newSequence, curNucleotide)
	}

	return newSequence, nil
}

func DNASequences(uncheckedSequences []string) ([]dnaSequence, error) {
	sequences := []dnaSequence{}
	for _, unchecked := range uncheckedSequences {
		checked, err := CreateDNASequence(unchecked)
		if err != nil {
			return sequences, err
		}

		sequences = append(sequences, checked)
	}

	return sequences, nil
}

func DNASequencesFromNucleotides(nucleotides []nucleotide) []dnaSequence {
	sequence := []dnaSequence{}
	for _, nucleotide := range nucleotides {
		sequence = append(sequence, dnaSequence{nucleotide})
	}
	return sequence
}

// The hamming distance is the number of differences of two dnaSequences
func (sequence dnaSequence) HammingDistance(compareSequence dnaSequence) int {
	mismatches := 0
	var shorterSequence dnaSequence
	var longerSequence dnaSequence
	if len(sequence) > len(compareSequence) {
		shorterSequence = compareSequence
		longerSequence = sequence
	} else {
		shorterSequence = sequence
		longerSequence = compareSequence
	}

	for i, value := range shorterSequence {
		if value != longerSequence[i] {
			mismatches++
		}
	}

	mismatches += len(longerSequence) - len(shorterSequence)
	return mismatches
}

// Finds patterns similar within a tolerance and returns their starting indexes
func SimilarPatterns(genome dnaSequence, pattern dnaSequence, tolerance int) []int {
	patternIndexes := []int{}
	for i := 0; i <= len(genome)-len(pattern); i++ {
		endIndex := min(i+len(pattern), len(genome))
		comparePattern := genome[i:endIndex]
		if len(comparePattern) < len(pattern) && i < len(pattern) {
			comparePattern = LeftPad(comparePattern, len(pattern)-len(comparePattern))
		}
		if comparePattern.HammingDistance(pattern) <= tolerance {
			patternIndexes = append(patternIndexes, i)
		}
	}

	return patternIndexes
}

// Left pad for Hamming distance, pads with bad nucleotides
func LeftPad(toPad dnaSequence, padLength int) dnaSequence {
	padding := make([]nucleotide, 0, padLength)
	for i := 0; i < padLength; i++ {
		padding = append(padding, nucleotide(PadNucleotide()))
	}

	return append(padding, toPad...)
}

func (genome dnaSequence) PatternCount(pattern dnaSequence) int {
	var count int = 0
	var index int = 0
	for index <= len(genome)-len(pattern) {
		if reflect.DeepEqual(genome[index:index+len(pattern)], pattern) {
			count++
		}
		index++
	}

	return count
}

func (genome dnaSequence) PatternCountWithMismatches(pattern dnaSequence, tolerance int) int {
	return len(SimilarPatterns(genome, pattern, tolerance))
}

func (sequence dnaSequence) String() string {
	byteSequence := []byte{}
	for _, nucleotide := range sequence {
		byteSequence = append(byteSequence, byte(nucleotide))
	}
	return string(byteSequence)
}

// The `5 to `3 skew or G to C skew
func Skew(genome dnaSequence) []int {
	var skews []int = make([]int, len(genome)+1)
	var currentSkew int = 0

	skews[0] = currentSkew
	for i, curNucleotide := range genome {
		if curNucleotide == 'G' {
			currentSkew++
		} else if curNucleotide == 'C' {
			currentSkew--
		}
		skews[i+1] = currentSkew
	}

	return skews
}

// A neighbor  are kmers that are within a Hamming Distance away from the given
// pattern.
func GenerateNeighbors(pattern dnaSequence, distance int) (neighborhood []dnaSequence) {
	neighborhood = []dnaSequence{}
	if distance == 0 {
		neighborhood = []dnaSequence{pattern}
		return
	}
	if len(pattern) == 1 {
		neighborhood = DNASequencesFromNucleotides(GetValidNucleotidesSlice())
		return
	}

	suffix := pattern[1:]
	suffixNeighbors := GenerateNeighbors(suffix, distance)
	for _, sequence := range suffixNeighbors {
		// If the hamming distance allows a mismatch, add all possible mismatches
		if suffix.HammingDistance(sequence) < distance {
			for _, nucl := range GetValidNucleotides() {
				newSequence := append(dnaSequence{nucl}, sequence...)
				neighborhood = append(neighborhood, newSequence)
			}
		} else {
			firstNucleotideSequence := dnaSequence{pattern[0]}
			newSequence := append(firstNucleotideSequence, sequence...)
			neighborhood = append(neighborhood, newSequence)
		}
	}

	return
}

/*
func MostFrequentKmersWithMismatch(genome dnaSequence, k int, tolerance int) []dnaSequence {
	allKmers := FindAllKmers(genome, k)
	maxCountKmers := []string{}
	maxCount := 0
	for _, kmer := range allKmers {
		similarPatternIndexes := SimilarPatterns(text, kmer, tolerance)
		if len(similarPatternIndexes) > maxCount {
			maxCountKmers = []string{kmer}
			// Retrieve all similar kmers.
			maxCountKmers = append(maxCountKmers,
				RetrieveKmersFromIndexSlice(text, similarPatternIndexes, k)...)
			maxCount = len(similarPatternIndexes)
		} else if len(similarPatternIndexes) == maxCount {
			maxCountKmers = append(maxCountKmers, kmer)
		}
	}

	return RemoveDuplicates(maxCountKmers)
}
*/

func FrequentKmersWithMismatches(genome dnaSequence, k int, tolerance int) []dnaSequence {
	mers := FindAllKmers(genome, k)
	var mostFrequent []dnaSequence
	highestFrequency := 0
	for _, mer := range mers {
		neighbors := GenerateNeighbors(mer, tolerance)
		for _, neighbor := range neighbors {
			currentFrequency := genome.PatternCountWithMismatches(neighbor, tolerance)
			if highestFrequency < currentFrequency {
				mostFrequent = []dnaSequence{neighbor}
				mostFrequent = append(mostFrequent, neighbor)
				highestFrequency = currentFrequency
			} else if highestFrequency == currentFrequency {
				mostFrequent = append(mostFrequent, neighbor)
			}
		}
	}

	return RemoveDuplicates(mostFrequent)
}

func FrequentKmersWithMismatchesAndReverseComplements(genome dnaSequence, k int, tolerance int) []dnaSequence {
	mers := FindAllKmers(genome, k)
	var mostFrequent []dnaSequence
	highestFrequency := 0
	for _, mer := range mers {
		neighbors := GenerateNeighbors(mer, tolerance)
		for _, neighbor := range neighbors {
			reverseComplement := neighbor.GenerateReverseComplement()
			currentFrequency := genome.PatternCountWithMismatches(neighbor, tolerance)
			currentFrequency += genome.PatternCountWithMismatches(reverseComplement, tolerance)
			if highestFrequency < currentFrequency {
				mostFrequent = []dnaSequence{neighbor}
				mostFrequent = append(mostFrequent, neighbor)
				mostFrequent = append(mostFrequent, reverseComplement)
				highestFrequency = currentFrequency
			} else if highestFrequency == currentFrequency {
				mostFrequent = append(mostFrequent, neighbor)
				mostFrequent = append(mostFrequent, reverseComplement)
			}
		}
	}

	return RemoveDuplicates(mostFrequent)
}

func (sequence dnaSequence) GenerateReverseComplement() dnaSequence {
	reverse := dnaSequence{}
	for i := len(sequence) - 1; i >= 0; i-- {
		reverse = append(reverse, sequence[i].Complement())
	}
	return reverse
}

func FindAllKmers(genome dnaSequence, k int) []dnaSequence {
	kmers := []dnaSequence{}
	for i := 0; i <= len(genome)-k; i++ {
		kmers = append(kmers, genome[i:i+k])
	}
	return RemoveDuplicates(kmers)
}

func RemoveDuplicates(mers sequenceGroup) []dnaSequence {
	for i := 0; i < len(mers)-1; i++ {
		currentMer := mers[i]
		for compareIndex := i + 1; compareIndex < len(mers); compareIndex++ {
			compareMer := mers[compareIndex]
			if currentMer.Equals(compareMer) {
				mers = mers.Remove(compareIndex)
				compareIndex--
			}
		}
	}

	return mers
}

// Motifs are patterns that are present accross multiple dna sequences.
//
// FindMotifs returns the maximized result of all motifs that are k length
// and d mismatches from eachother.
func FindMotifs(sequences []dnaSequence, k int, d int) []dnaSequence {
	motifs := []dnaSequence{}
	allKmers := []dnaSequence{}
	for _, mer := range sequences {
		allKmers = append(allKmers, FindAllKmers(mer, k)...)
	}

	for _, currentKmer := range allKmers {
		for _, generated := range GenerateNeighbors(currentKmer, d) {
			if AllContain(sequences, generated, d) {
				motifs = append(motifs, generated)
			}
		}
	}

	return RemoveDuplicates(motifs)
}

func AllContain(haystacks []dnaSequence, needle dnaSequence, d int) bool {
	for _, sequence := range haystacks {
		if !sequence.Contains(needle, d) {
			return false
		}
	}

	return true
}

func (haystack dnaSequence) Contains(needle dnaSequence, d int) bool {
	for i := 0; i <= len(haystack)-len(needle); i++ {
		curSequence := haystack[i:min(i+len(needle), len(haystack))]
		if curSequence.HammingDistance(needle) <= d {
			return true
		}
	}

	return false
}

// Compares two sequences, and returns whether they are the same.
func (sequenceA dnaSequence) Equals(sequenceB dnaSequence) bool {
	if len(sequenceA) != len(sequenceB) {
		return false
	}
	for i := 0; i < len(sequenceA); i++ {
		if sequenceA[i] != sequenceB[i] {
			return false
		}
	}

	return true
}

/*
TODO: Implement quick sorting.
func (group sequenceGroup) Sort() {
	for
	group.
}

func (group sequenceGroup) partition() {
}
*/

func (sequence dnaSequence) String() string {
	return sequence
}

func (sequence dnaSequence) Compare(compareSequence dnaSequence) int {
	return strings.Compare(string([]nucleotide(sequence)), string([]nucleotide(compareSequence)))
}

func (group sequenceGroup) Sort() sequenceGroup {
	for i := 0; i < len(group)-1; i++ {
		lowest := i
		for k := i + 1; k < len(group); k++ {
			if group[lowest].Compare(group[k]) == 1 {
				lowest = k
			}
		}

		if lowest != i {
			tmp := group[i]
			group[i] = group[lowest]
			group[lowest] = tmp
		}
	}

	return group
}

func (group sequenceGroup) FindSequence(needle dnaSequence) int {
	for index, curSequence := range group {
		if curSequence.Equals(needle) {
			return index
		}
	}

	return -1
}

func (groupA sequenceGroup) Diff(groupB sequenceGroup) (diffA sequenceGroup, diffB sequenceGroup) {
	groupA = RemoveDuplicates(groupA)
	groupB = RemoveDuplicates(groupB)

	for i := 0; i < len(groupA); i++ {
		if searchIndex := groupB.FindSequence(groupA[i]); searchIndex != -1 {
			groupA = groupA.Remove(i)
			groupB = groupB.Remove(searchIndex)
			i--
		}
	}

	for i := 0; i < len(groupB); i++ {
		if searchIndex := groupA.FindSequence(groupB[i]); searchIndex != -1 {
			groupA = groupA.Remove(searchIndex)
			groupB = groupB.Remove(i)
			i--
		}
	}

	return groupA, groupB
}

// Removes index i from the group.
func (group sequenceGroup) Remove(i int) sequenceGroup {
	group = append(group[:i], group[i+1:]...)

	return group
}
