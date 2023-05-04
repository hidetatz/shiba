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

		// comment
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

		// number
		if (isdigit(c)) {
			bool decimals = false;
			std::string s = "";
			while (true) {
				if (isdigit(c)) {
					s.push_back(c);
					c++;
				} else if (c == '.') {
					decimals = true;
					s.push_back(c);
					c++;
				} else {
					break;
				}
			}
			if (decimals) {

			} else {
				int i = std::stoi(s);
				Token t{TokenType::_int, msg, 0, 0};
				tokens.push_back(t);
			}

			Token t{TokenType::comment, msg, 0, 0};
			tokens.push_back(t);
			continue;
		}
	}

	return tokens;
}

} // namespace shiba
