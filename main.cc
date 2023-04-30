#include <iostream>
#include <fstream>
#include <string>

void run_repl() {
	std::ifstream infile("main.f");

	std::string line;
	while (std::getline(infile, line))
	{
		
		std::istringstream iss(line);
		int a, b;
		if (!(iss >> a >> b)) { break; } // error
		
		// process pair (a,b)
	}

}

int main() {
	std::cout << "Now shiba lang is up and running..." << std::endl;
	run_repl();
	return 0;
}
