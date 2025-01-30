package rag_test

import (
	"fmt"
	"testing"

	"github.com/jtarchie/builder/rag"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

const doc1 = `
# Document 1: Introduction to Machine Learning

## What is Machine Learning?
Machine learning is a subset of artificial intelligence (AI) that enables systems to learn and improve from experience without being explicitly programmed.

## Types of Machine Learning
- **Supervised Learning**: The model learns from labeled data.
- **Unsupervised Learning**: The model finds patterns in unlabeled data.
- **Reinforcement Learning**: The model learns by interacting with an environment and receiving rewards.

## Applications of Machine Learning
Machine learning is used in various fields, including:
- Healthcare (disease prediction, personalized medicine)
- Finance (fraud detection, stock market predictions)
- Natural Language Processing (chatbots, sentiment analysis)
`

const doc2 = `
# Document 2: The Solar System

## Overview
The Solar System consists of the Sun and the celestial bodies that orbit it, including planets, moons, asteroids, and comets.

## The Planets
The eight planets of the Solar System are:
1. Mercury
2. Venus
3. Earth
4. Mars
5. Jupiter
6. Saturn
7. Uranus
8. Neptune

## Interesting Facts
- Jupiter is the largest planet.
- Venus has the hottest surface temperature.
- Earth is the only known planet with life.
`

const doc3 = `
# Document 3: The History of the Internet

## Early Beginnings
The concept of a global computer network was first envisioned in the 1960s. The ARPANET, a precursor to the modern Internet, was developed by the U.S. Department of Defense.

## Key Milestones
- **1969**: ARPANET goes live.
- **1983**: Introduction of the TCP/IP protocol, forming the foundation of the modern Internet.
- **1991**: The World Wide Web (WWW) is introduced by Tim Berners-Lee.
- **2000s**: Rise of social media, e-commerce, and cloud computing.

## Impact on Society
The Internet has transformed:
- Communication (email, messaging apps)
- Commerce (online shopping, digital banking)
- Education (e-learning, online courses)
`

func TestRAG(t *testing.T) {
	assert := gomega.NewWithT(t)

	config := &rag.OpenAIConfig{
		EmbedModel: "nomic-embed-text",
		Endpoint:   "http://localhost:11434/v1",
		LLMModel:   "llama3.2",
		Token:      "",
	}
	rag, err := rag.New(":memory:", config)
	assert.Expect(err).NotTo(HaveOccurred())

	for index, doc := range []string{doc1, doc2, doc3} {
		err = rag.AddDocument(fmt.Sprintf("id-%d", index), doc)
		assert.Expect(err).NotTo(HaveOccurred())
	}

	results, err := rag.Search("machine learning")
	assert.Expect(err).NotTo(HaveOccurred())
	assert.Expect(results).To(HaveLen(1))

	answer, err := rag.Ask("What is the largest planet?")
	assert.Expect(err).NotTo(HaveOccurred())
	fmt.Println("answer:" + answer)
}
