#include <ctype.h>
#include <vector>
#include <string>
#include <iostream>

#include "shiba.h"

namespace shiba {

std::vector<Token> tokenize(std::string line) {
	std::vector<Token> tokens;

	std::cout << line << std::endl;
	for(char &c : line) {
		// skip whitespace characters
		if (isspace(c)) {
			continue;
		}

		// terminate tokenize if it's a line comment
		if (c == '#') {
			c++;
			// skip whitespaces after hash
			while (isspace(c) || c == '#') {
				c++;
			}
			std::string msg = "";
			while(c != '\n') {
				msg.push_back(c);
				c++;
			}
			Token t{TokenType::comment, msg, 0, 0};
			tokens.push_back(t);
			break;
		}
	}

	return tokens;
}

} // namespace shiba
