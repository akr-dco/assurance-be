pipeline {
    agent any

    environment {
        TARGET_USER = "docker_prod"
        TARGET_HOST = "192.168.150.150"
    }

    stages {

        stage('Resolve Target Directory') {
            steps {
                script {
                    if (env.BRANCH_NAME == 'main') {
                        env.TARGET_DIR = "/home/docker_prod/cicd-prod/assurance-be"
                    } else if (env.BRANCH_NAME == 'dev') {
                        env.TARGET_DIR = "/home/docker_prod/cicd-dev/assurance-be"
                    } else {
                        error "Branch ${env.BRANCH_NAME} is not allowed to deploy"
                    }
                }

                echo "Branch        : ${env.BRANCH_NAME}"
                echo "Deploy target : ${TARGET_USER}@${TARGET_HOST}:${env.TARGET_DIR}"
            }
        }

        stage('Prepare Target') {
            steps {
                sshagent(['privatekey-akr']) {
                    sh """
                    ssh -o StrictHostKeyChecking=no ${TARGET_USER}@${TARGET_HOST} '
                        mkdir -p ${TARGET_DIR}
                    '
                    """
                }
            }
        }

        stage('Sync Repository') {
            steps {
                sshagent(['privatekey-akr']) {
                    sh """
                    rsync -avz --delete \
                        --exclude '.git' \
                        --exclude '.jenkins' \
                        ./ \
                        ${TARGET_USER}@${TARGET_HOST}:${TARGET_DIR}/
                    """
                }
            }
        }

        stage('Deploy Docker Compose') {
            steps {
                sshagent(['privatekey-akr']) {
                    sh """
                    ssh -o StrictHostKeyChecking=no ${TARGET_USER}@${TARGET_HOST} '
                        cd ${TARGET_DIR} &&
                        docker compose down &&
                        docker compose build &&
                        docker compose up -d
                    '
                    """
                }
            }
        }
    }

    post {
        success {
            echo "✅ Deployment SUCCESS for ${env.BRANCH_NAME}"
        }
        failure {
            echo "❌ Deployment FAILED for ${env.BRANCH_NAME}"
        }
    }
}
