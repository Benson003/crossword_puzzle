#!/usr/bin/python3
import json
import os
import collections

def generate_optimized_dictionary():
    dict_path = "/usr/share/dict/words"
    if not os.path.exists(dict_path):
        return {"3": ["ZIG", "GPU"], "4": ["ARCH", "NIRI"]}

    word_map = {}
    
    # Crossword "Power Letters" - words with these intersect easily
    power_letters = set("RSTLNATEIO") 

    with open(dict_path, 'r') as f:
        all_words = [line.strip().upper() for line in f]

    # Pre-filter for valid crossword words
    valid_words = [w for w in all_words if w.isalpha() and 3 <= len(w) <= 15]

    for word in valid_words:
        length_str = str(len(word))
        if length_str not in word_map:
            word_map[length_str] = []
        
        # Scoring: How "flexible" is this word for intersections?
        # More vowels and power consonants = Higher score
        score = sum(1 for char in word if char in power_letters)
        
        # We store the word and its score as a tuple temporarily
        word_map[length_str].append((word, score))

    final_map = {}
    for length, items in word_map.items():
        # Sort words by their score (descending) so the solver tries 
        # "flexible" words like "ESTATE" before "ZYGOTE"
        items.sort(key=lambda x: x[1], reverse=True)
        
        # Take the top 3000 most "flexible" words per length
        # This keeps the dictionary high-quality and fast
        final_map[length] = [w[0] for w in items[:3000]]

    return final_map

if __name__ == "__main__":
    words = generate_optimized_dictionary()
    with open("dictionary.json", "w") as f:
        json.dump(words, f, indent=2)
    
    print(f"[SUCCESS] Generated high-flexibility dictionary.")
    for length in sorted(words.keys(), key=int):
        print(f"Length {length}: {len(words[length])} words")
