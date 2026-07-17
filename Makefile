####################################################### VARIABLE #######################################################
NAME		:=	hypertube
SRCS_D		:=	srcs
ENV_EXEMPLE	:=	.env.exemple
DATA_DIR	:=	$(SRCS_D)/data
ENV_FILE	:=	$(SRCS_D)/.env
COMPOSE_F	:=	$(SRCS_D)/docker-compose.yml
SERVICE		?=	#Leave blank

######################################################## FLAGS #########################################################
FLAGS		=	-f $(COMPOSE_F)
COMPOSE		=	docker compose
DSHELL		=	/bin/sh

######################################################## RULES #########################################################
.DEFAULT_GOAL = all

.PHONY: all
all			:	$(NAME)

$(NAME)		:
			mkdir -p $(DATA_DIR)
			$(COMPOSE) $(FLAGS) up --build $(SERVICE)

CMDS		:=	up build down ps ls images events top
.PHONY: $(CMDS)
$(CMDS)		:
			$(COMPOSE) $(FLAGS) $@ $(SERVICE)

.PHONY: detach
detach		:
			$(COMPOSE) $(FLAGS) up --$@ $(SERVICE)

.PHONY: logs
logs		:	build
			$(COMPOSE) $(FLAGS) $@ -f $(SERVICE)

.PHONY: exec
exec		:
			$(COMPOSE) $(FLAGS) $@ $(SERVICE) $(DSHELL)

.PHONY: env
env			:
			./launch.d/01generatePasswordsAndKeys.sh

.PHONY: clean
clean		:
			$(COMPOSE) $(FLAGS) down --rmi local --remove-orphans

.PHONY: vclean
vclean		:
			$(COMPOSE) $(FLAGS) down -v --remove-orphans
			rm -rf $(ENV_FILE)
			rm -rf $(DATA_DIR)

PHONY: fclean
fclean		:
			$(COMPOSE) $(FLAGS) down -v --rmi all --remove-orphans
			rm -rf $(ENV_FILE)
			rm -rf $(DATA_DIR)

.PHONY: image-ls image-rm
image-ls	:
			docker image ls -a
image-rm	:
			docker image rm `docker image ls -qa`

.PHONY: container-ls container-rm
container-ls:
			docker container ls -a
container-rm:
			docker container rm `docker container ls -qa`

.PHONY: volume-ls volume-rm
volume-ls	:
			docker volume ls
volume-rm	:
			docker volume rm `docker volume ls -qa`

.PHONY: network-ls network-rm
network-ls	:
			docker network ls
network-rm	:
			docker network rm `docker network ls -qa`

.PHONY: prune
prune		:
			docker system prune -af

.PHONY: sre
sre			:	clean all

.PHONY: vre
vre			:	vclean all

.PHONY: re
re			:	fclean env all
