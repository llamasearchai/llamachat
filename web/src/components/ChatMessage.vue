<template>
  <div 
    class="chat-message" 
    :class="{ 
      'own-message': isOwnMessage,
      'ai-message': message.isAiGenerated,
      'deleted': message.isDeleted
    }"
  >
    <div class="message-avatar" v-if="!isOwnMessage">
      <img 
        v-if="message.user && message.user.avatarUrl" 
        :src="message.user.avatarUrl" 
        :alt="message.user.displayName || message.user.username" 
      />
      <div v-else class="avatar-placeholder">
        {{ userInitial }}
      </div>
    </div>
    
    <div class="message-content">
      <div class="message-header">
        <span class="username" v-if="!isOwnMessage && message.user">
          {{ message.user.displayName || message.user.username }}
        </span>
        <span class="timestamp">
          {{ formattedTime }}
        </span>
        <div class="message-status" v-if="isOwnMessage">
          <span v-if="message.isEdited" class="edited-indicator">edited</span>
          <span class="status-icon" :class="deliveryStatus"></span>
        </div>
      </div>
      
      <div class="message-body">
        <p v-if="message.isDeleted" class="deleted-message">
          This message has been deleted
        </p>
        <p v-else-if="message.isEncrypted && !decrypted" class="encrypted-message">
          <lock-icon size="small" />
          <span>Encrypted message. <button @click="decryptMessage">Decrypt</button></span>
        </p>
        <div v-else v-html="formattedContent"></div>
      </div>
      
      <div class="message-actions" v-if="!message.isDeleted && showActions">
        <button @click="onReply" class="action-button">
          <reply-icon size="small" />
        </button>
        <button v-if="isOwnMessage" @click="onEdit" class="action-button">
          <edit-icon size="small" />
        </button>
        <button v-if="isOwnMessage || canModerate" @click="onDelete" class="action-button">
          <delete-icon size="small" />
        </button>
        <button @click="onCopy" class="action-button">
          <copy-icon size="small" />
        </button>
      </div>
      
      <div v-if="message.replyTo" class="replied-message">
        <div class="reply-indicator">
          <reply-icon size="small" />
          <span>Replying to {{ message.replyTo.user ? message.replyTo.user.displayName || message.replyTo.user.username : 'Unknown' }}</span>
        </div>
        <div class="reply-content">{{ truncateReply(message.replyTo.content) }}</div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, onMounted } from 'vue';
import DOMPurify from 'dompurify';
import marked from 'marked';
import { formatDistanceToNow } from 'date-fns';
import LockIcon from './icons/LockIcon.vue';
import ReplyIcon from './icons/ReplyIcon.vue';
import EditIcon from './icons/EditIcon.vue';
import DeleteIcon from './icons/DeleteIcon.vue';
import CopyIcon from './icons/CopyIcon.vue';

export default {
  name: 'ChatMessage',
  
  components: {
    LockIcon,
    ReplyIcon,
    EditIcon,
    DeleteIcon,
    CopyIcon
  },
  
  props: {
    message: {
      type: Object,
      required: true
    },
    currentUserId: {
      type: String,
      required: true
    },
    canModerate: {
      type: Boolean,
      default: false
    }
  },
  
  setup(props, { emit }) {
    const showActions = ref(false);
    const decrypted = ref(!props.message.isEncrypted);
    
    const isOwnMessage = computed(() => {
      return props.message.user && props.message.user.id === props.currentUserId;
    });
    
    const userInitial = computed(() => {
      if (!props.message.user) return '?';
      const name = props.message.user.displayName || props.message.user.username;
      return name.charAt(0).toUpperCase();
    });
    
    const formattedTime = computed(() => {
      try {
        const date = new Date(props.message.createdAt);
        return formatDistanceToNow(date, { addSuffix: true });
      } catch (e) {
        return 'Unknown time';
      }
    });
    
    const formattedContent = computed(() => {
      if (props.message.isDeleted) return '';
      
      // Use markdown for formatting
      let html = marked(props.message.content || '');
      
      // Sanitize HTML to prevent XSS
      html = DOMPurify.sanitize(html);
      
      return html;
    });
    
    const deliveryStatus = computed(() => {
      if (props.message.isDelivered) return 'delivered';
      if (props.message.isSent) return 'sent';
      return 'pending';
    });
    
    function truncateReply(content) {
      if (!content) return '';
      return content.length > 50 ? content.substring(0, 50) + '...' : content;
    }
    
    function decryptMessage() {
      // In a real implementation, this would decrypt the message
      // For this example, we'll just toggle the decrypted state
      decrypted.value = true;
    }
    
    function onReply() {
      emit('reply', props.message);
    }
    
    function onEdit() {
      emit('edit', props.message);
    }
    
    function onDelete() {
      emit('delete', props.message);
    }
    
    function onCopy() {
      if (navigator.clipboard && props.message.content) {
        navigator.clipboard.writeText(props.message.content)
          .then(() => {
            // Could show a toast notification that copying succeeded
          })
          .catch(err => {
            console.error('Failed to copy text: ', err);
          });
      }
    }
    
    onMounted(() => {
      // Auto-decrypt if the user has this preference enabled
      if (props.message.isEncrypted && !decrypted.value) {
        // Check user preferences and auto-decrypt if enabled
        // For this example, we'll just leave it as is
      }
    });
    
    return {
      showActions,
      decrypted,
      isOwnMessage,
      userInitial,
      formattedTime,
      formattedContent,
      deliveryStatus,
      truncateReply,
      decryptMessage,
      onReply,
      onEdit,
      onDelete,
      onCopy
    };
  }
};
</script>

<style scoped>
.chat-message {
  display: flex;
  margin-bottom: 16px;
  position: relative;
  transition: all 0.2s ease;
}

.chat-message:hover .message-actions {
  opacity: 1;
}

.own-message {
  flex-direction: row-reverse;
}

.own-message .message-content {
  background-color: var(--own-message-bg, #dcf8c6);
  border-radius: 12px 2px 12px 12px;
}

.own-message .message-header {
  flex-direction: row-reverse;
}

.message-avatar {
  width: 40px;
  height: 40px;
  border-radius: 50%;
  overflow: hidden;
  margin-right: 8px;
  flex-shrink: 0;
}

.own-message .message-avatar {
  margin-right: 0;
  margin-left: 8px;
}

.message-avatar img {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.avatar-placeholder {
  width: 100%;
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  background-color: var(--avatar-bg, #ccc);
  color: var(--avatar-text, #fff);
  font-weight: bold;
}

.message-content {
  max-width: 70%;
  background-color: var(--message-bg, #fff);
  border-radius: 2px 12px 12px 12px;
  padding: 8px 12px;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.1);
}

.ai-message .message-content {
  background-color: var(--ai-message-bg, #f0f0ff);
}

.message-header {
  display: flex;
  align-items: center;
  margin-bottom: 4px;
  font-size: 0.85rem;
}

.username {
  font-weight: bold;
  margin-right: 8px;
  color: var(--username-color, #333);
}

.timestamp {
  color: var(--timestamp-color, #999);
  font-size: 0.8rem;
}

.message-status {
  display: flex;
  align-items: center;
  margin-left: 8px;
}

.edited-indicator {
  font-size: 0.7rem;
  color: var(--edited-color, #999);
  margin-right: 4px;
}

.status-icon {
  width: 14px;
  height: 14px;
  display: inline-block;
}

.status-icon.delivered {
  color: var(--delivered-color, #4caf50);
}

.status-icon.sent {
  color: var(--sent-color, #2196f3);
}

.status-icon.pending {
  color: var(--pending-color, #9e9e9e);
}

.message-body {
  word-break: break-word;
  font-size: 0.95rem;
  line-height: 1.4;
  white-space: pre-wrap;
}

.deleted-message {
  font-style: italic;
  color: var(--deleted-color, #999);
}

.encrypted-message {
  display: flex;
  align-items: center;
  color: var(--encrypted-color, #777);
  font-size: 0.9rem;
}

.encrypted-message button {
  background: none;
  border: none;
  color: var(--link-color, #2196f3);
  text-decoration: underline;
  cursor: pointer;
  margin-left: 4px;
  font-size: 0.9rem;
}

.message-actions {
  position: absolute;
  right: 0;
  top: -20px;
  display: flex;
  background-color: var(--actions-bg, #fff);
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
  opacity: 0;
  transition: opacity 0.2s ease;
}

.own-message .message-actions {
  right: auto;
  left: 0;
}

.action-button {
  background: none;
  border: none;
  padding: 4px;
  cursor: pointer;
  color: var(--action-color, #555);
  transition: color 0.2s ease;
}

.action-button:hover {
  color: var(--action-hover-color, #000);
}

.replied-message {
  margin-top: 8px;
  padding-top: 4px;
  border-top: 1px solid var(--reply-border, #eee);
  font-size: 0.85rem;
}

.reply-indicator {
  display: flex;
  align-items: center;
  color: var(--reply-color, #666);
  margin-bottom: 2px;
}

.reply-content {
  color: var(--reply-content, #777);
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

/* Dark mode styles */
@media (prefers-color-scheme: dark) {
  .message-content {
    --message-bg: #2a2a2a;
    color: #eee;
  }
  
  .own-message .message-content {
    --own-message-bg: #2b5031;
  }
  
  .ai-message .message-content {
    --ai-message-bg: #252544;
  }
  
  .username {
    --username-color: #eee;
  }
  
  .timestamp {
    --timestamp-color: #aaa;
  }
  
  .message-actions {
    --actions-bg: #333;
  }
  
  .action-button {
    --action-color: #ccc;
    --action-hover-color: #fff;
  }
  
  .deleted-message {
    --deleted-color: #999;
  }
  
  .encrypted-message {
    --encrypted-color: #aaa;
  }
  
  .encrypted-message button {
    --link-color: #90caf9;
  }
  
  .replied-message {
    --reply-border: #444;
  }
  
  .reply-indicator {
    --reply-color: #aaa;
  }
  
  .reply-content {
    --reply-content: #bbb;
  }
}
</style> 