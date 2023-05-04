#include <iostream>
#include <fstream>
#include <string>
#include <vector>

#include "shiba.h"

void run(char* filename) {
	std::ifstream infile(filename);

	std::string line;
	while (std::getline(infile, line)) {
		std::vector<shiba::Token> tokens = shiba::tokenize(line);
		std::cout << tokens.size() << std::endl;
	}
}


int main(int argc, char* argv[]) {
	if (argc != 2) {
		std::cout << "a filename to be run must be given!" << std::endl;
		exit(1);
	}

	run(argv[1]);
	return 0;
}
