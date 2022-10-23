package main

/*
  #include <stdio.h>
  #include <unistd.h>
  #include <termios.h>

  struct termios old = {0};

  char getch(){
      char ch = 0;
      // struct termios old = {0};
      fflush(stdout);
      if( tcgetattr(0, &old) < 0 ) perror("tcsetattr()");
      old.c_lflag &= ~ICANON;
      old.c_lflag &= ~ECHO;
      old.c_cc[VMIN] = 1;
      old.c_cc[VTIME] = 0;
      if( tcsetattr(0, TCSANOW, &old) < 0 ) perror("tcsetattr ICANON");
      if( read(0, &ch,1) < 0 ) perror("read()");
      old.c_lflag |= ICANON;
      old.c_lflag |= ECHO;
      if(tcsetattr(0, TCSADRAIN, &old) < 0) perror("tcsetattr ~ICANON");
      return ch;
  }

  void clean_up(){
      old.c_lflag |= ICANON;
      old.c_lflag |= ECHO;
      if(tcsetattr(0, TCSADRAIN, &old) < 0) perror("tcsetattr ~ICANON");
  }

*/
import "C"

// For tests only
func termiosStart(channelStop chan int) {
	go func() {
		for {
			key := C.getch()
			if key == 27 {
				channelStop <- 1
				stopPhone()
			}
		}
	}()
}

func termiosCleanUp() {
	C.clean_up()
}
