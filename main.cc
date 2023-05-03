#include <iostream>
#include <fstream>
#include <string>
#include <vector>

#include "shiba.h"

void run_repl() {
	std::ifstream infile("test.sb");

	std::string line;
	while (std::getline(infile, line)) {
		std::cout << line << std::endl;
		std::vector<shiba::Token> tokens = shiba::tokenize(line);
		std::cout << tokens.size() << std::endl;
	}
}

int main() {
	std::cout << "Now shiba lang is up and running..." << std::endl;
	run_repl();
	return 0;
}
