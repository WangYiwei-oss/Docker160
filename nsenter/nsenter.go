package nsenter

/*
  #define _GNU_SOURCE
  #include <stdlib.h>
  #include <errno.h>
  #include <stdio.h>
  #include <string.h>
  #include <fcntl.h>
  #include <sched.h>
  #include <unistd.h>

//__attribute__((constructor))指的是一旦这个包被引用，那么这个函数就会被自动执行
//类似于构造函数，会在程序一启动的时候执行
__attribute__((constructor)) void enter_namespace(void){
          char *mydocker_pid;
          //从环境变量中获取需要进入的PID
          mydocker_pid = getenv("mydocker_pid");
          if (mydocker_pid){
                  //注意这里是通过环境变量来控制要不要执行进入的操作的
                  fprintf(stdout,"got mydocker_pid=%s\n",mydocker_pid);
          }else{
                  fprintf(stdout,"missing mydocker_pid env,so skip nsenter");
                  return;
          }
          char *mydocker_cmd;
          //从环境变量里面获取需要执行的命令
          mydocker_cmd=getenv("mydocker_cmd");
          if (mydocker_cmd){
                  fprintf(stdout,"got mydocker_cmd=%s\n",mydocker_cmd);
          }else{
                  fprintf(stdout,"missing mydocker_cmd env skip nsenter");
                  return;
          }
          int i;
          char nspath[1024];
          char *namespaces[]={"ipc","uts","net","pid","mnt"};
          for (i=0;i<5;i++){
                  //拼接对应的路径/proc/pid/ns/ipc类似这样
                  sprintf(nspath,"/proc/%s/ns/%s",mydocker_pid,namespaces[i]);
                  int fd = open(nspath,O_RDONLY);
                  //这里调用setns系统调用进入对应的Namespace
		  if (setns(fd,0)==-1){
                          fprintf(stderr,"setns on %s namespace failed: %s\n",namespaces[i],strerror(errno));
                  }else{
                          fprintf(stdout,"setns succeeded");
                  }
                close(fd);
          }
          //在进入的Namespace中执行指定的命令
          int res = system(mydocker_cmd);
          exit(0);
          return;
  }
*/
import "C"
