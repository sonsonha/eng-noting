package ai

const systemPrompt = `
You are an English teacher for non-native learners.
Your explanations must be:
- Simple
- Clear
- Accurate
- Suitable for CEFR A2–B1 learners

Rules:
- Use simple English only
- Do NOT use the target word in the definition
- Explain only the most common meaning
- Avoid idioms and rare usages
- Keep sentences short
`

func explanationPrompt(word, context string) string {
	return `
Word: "` + word + `"
Context sentence (if any): "` + context + `"

Task:
1. Give a simple definition
2. Give ONE correct example sentence
3. Give ONE incorrect or unnatural example sentence
4. State the part of speech
5. Guess CEFR level (A2, B1, or B2)

Output in JSON only.
`
}
