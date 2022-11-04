@Library('devops-lib') _

pipeline {
    agent any
    parameters {
        string(name: 'SERVICE', defaultValue: 'media-server', description: 'Service what will be build and deploy')
        string(name: 'DOCKER_REGISTRY', defaultValue: 'nexus.optmoskva.ru:8444/tcm/media-server', description: 'Internal NEXUS docker registry URL')
        string(name: 'IMAGE_TAG', defaultValue: '', description: 'Docker tag for manual deploy')
        choice(name: 'K8S_NS', choices: ['dev', 'prod'],description: 'Environment for deploy')
    }
	options {
		buildDiscarder(logRotator(numToKeepStr: '10', artifactNumToKeepStr: '10'))
		timestamps()
	}
    stages {
        stage ('Checkout and create docker tag from git commit'){
            steps   {
                checkout([$class: 'GitSCM', 
                        branches: [[name: '*/*']],
                        userRemoteConfigs: [[url: 'git@github.com:tcmoscow/media-server.git', credentialsId: 'git-mediaServer']]])
                sh 'git rev-parse HEAD > temp_hash'
                script { 
                    if ( (params.IMAGE_TAG).length() > 0 ) {
                        env.DOCKER_TAG = params.IMAGE_TAG
                    } else {
                        def commit = readFile('temp_hash').trim()
                        env.DOCKER_TAG = commit.substring(0,11).trim()
                    }
                }
                echo "Docker tag will be: ${env.DOCKER_TAG}"
                buildName "#${BUILD_NUMBER}#${K8S_NS}#${SERVICE}#${env.DOCKER_TAG}"
            }
        }
        stage ('Docker build'){
            when {
                expression { (params.IMAGE_TAG).length() == 0 }
                environment name:'K8S_NS', value:'dev'
            }
            steps {
                echo "Start building docker image for \"${params.SERVICE}\""
                echo "Docker tag will be: ${env.DOCKER_TAG}"
                sh "docker build -t ${DOCKER_REGISTRY}/${SERVICE}:${env.DOCKER_TAG} . "
                
            }
        }
        stage ('Docker push'){
            when {
                expression { (params.IMAGE_TAG).length() == 0 }
                environment name:'K8S_NS', value:'dev'
            }
            steps {
                echo "Start pushing docker image to \"${DOCKER_REGISTRY}\""
                withDockerRegistry(credentialsId: 'tyak-docker-user', url: "https://${DOCKER_REGISTRY}"){
                    sh "docker push ${DOCKER_REGISTRY}/${SERVICE}:${env.DOCKER_TAG}"
                }
            }
        }

        stage ('Docker Inspect'){
            when {
                anyOf {
                    expression { (params.IMAGE_TAG).length() != 0 }
                    environment name:'K8S_NS', value:'prod'
                }
            }
            steps {
                echo "Checking docker image exist: \"${DOCKER_REGISTRY}/${SERVICE}:${env.DOCKER_TAG}\""
                withDockerRegistry(credentialsId: 'tyak-docker-user', url: "https://${DOCKER_REGISTRY}"){
                    sh "docker image inspect ${DOCKER_REGISTRY}/${SERVICE}:${env.DOCKER_TAG} 1> /dev/null && echo \"image: ${DOCKER_REGISTRY}/${SERVICE}:${env.DOCKER_TAG} exist\""
                }
            }
        }
        stage ('Helm deploy'){
            steps {
                    deployHelm("${env.DOCKER_TAG}","$K8S_NS","$SERVICE")
            }
        }
    }
}