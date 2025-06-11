<template>
  <div class="monad-container">
    <div class="monad-core">
      <div class="perception-layer layer-1">
        <div class="perception-dot dot-1"></div>
        <div class="perception-dot dot-2"></div>
        <div class="perception-dot dot-3"></div>
        <div class="perception-dot dot-4"></div>
        <div class="perception-dot dot-5"></div>
        <div class="perception-dot dot-6"></div>
      </div>
      <div class="perception-layer layer-2">
        <div class="perception-dot dot-7"></div>
        <div class="perception-dot dot-8"></div>
        <div class="perception-dot dot-9"></div>
        <div class="perception-dot dot-10"></div>
      </div>
      <div class="perception-layer layer-3">
        <div class="perception-dot dot-11"></div>
        <div class="perception-dot dot-12"></div>
      </div>
      <div class="inner-essence"></div>
      <div class="consciousness-ripple ripple-1"></div>
      <div class="consciousness-ripple ripple-2"></div>
      <div class="consciousness-ripple ripple-3"></div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'MonadAnimation',
  props: {
    size: {
      type: String,
      default: '200px'
    }
  }
}
</script>

<style scoped>
.monad-container {
  display: flex;
  justify-content: center;
  align-items: center;
  width: v-bind(size);
  height: v-bind(size);
  position: relative;
  perspective: 1000px;
}

.monad-core {
  width: 120px;
  height: 120px;
  border-radius: 50%;
  background: transparent;
  position: relative;
  display: flex;
  justify-content: center;
  align-items: center;
  box-shadow: 
    0 0 30px rgba(180, 180, 180, 0.6),
    inset 0 0 20px rgba(255, 255, 255, 0.1);
  animation: monad-breathe 6s ease-in-out infinite;
  transform-style: preserve-3d;
}

.inner-essence {
  position: absolute;
  width: 20px;
  height: 20px;
  background: radial-gradient(circle, rgba(255, 255, 255, 1), rgba(220, 220, 220, 1));
  border-radius: 50%;
  box-shadow: 0 0 15px rgba(200, 200, 200, 0.8);
  animation: essence-pulse 3s ease-in-out infinite;
  z-index: 10;
}

.perception-layer {
  position: absolute;
  border-radius: 50%;
  border: 1px solid rgba(255, 255, 255, 0.3);
  animation: layer-rotate 8s linear infinite;
}

.layer-1 {
  width: 90px;
  height: 90px;
  animation-duration: 12s;
}

.layer-2 {
  width: 70px;
  height: 70px;
  animation-duration: 8s;
  animation-direction: reverse;
}

.layer-3 {
  width: 50px;
  height: 50px;
  animation-duration: 15s;
}

.perception-dot {
  position: absolute;
  width: 4px;
  height: 4px;
  background: rgba(255, 255, 255, 0.9);
  border-radius: 50%;
  box-shadow: 0 0 6px rgba(255, 255, 255, 0.8);
  animation: perception-twinkle 2s ease-in-out infinite;
}

.dot-1 { top: 5px; left: 50%; transform: translateX(-50%); animation-delay: 0s; }
.dot-2 { top: 50%; right: 5px; transform: translateY(-50%); animation-delay: 0.3s; }
.dot-3 { bottom: 5px; left: 50%; transform: translateX(-50%); animation-delay: 0.6s; }
.dot-4 { top: 50%; left: 5px; transform: translateY(-50%); animation-delay: 0.9s; }
.dot-5 { top: 20%; right: 20%; animation-delay: 1.2s; }
.dot-6 { bottom: 20%; left: 20%; animation-delay: 1.5s; }

.dot-7 { top: 8px; left: 50%; transform: translateX(-50%); animation-delay: 0.2s; }
.dot-8 { top: 50%; right: 8px; transform: translateY(-50%); animation-delay: 0.5s; }
.dot-9 { bottom: 8px; left: 50%; transform: translateX(-50%); animation-delay: 0.8s; }
.dot-10 { top: 50%; left: 8px; transform: translateY(-50%); animation-delay: 1.1s; }

.dot-11 { top: 12px; left: 50%; transform: translateX(-50%); animation-delay: 0.4s; }
.dot-12 { bottom: 12px; left: 50%; transform: translateX(-50%); animation-delay: 1.0s; }

.consciousness-ripple {
  position: absolute;
  border: 2px solid transparent;
  border-radius: 50%;
  animation: consciousness-expand 4s ease-out infinite;
}

.ripple-1 {
  width: 140px;
  height: 140px;
  animation-delay: 0s;
}

.ripple-2 {
  width: 160px;
  height: 160px;
  animation-delay: 1.3s;
}

.ripple-3 {
  width: 180px;
  height: 180px;
  animation-delay: 2.6s;
}

@keyframes monad-breathe {
  0%, 100% {
    transform: scale(1) rotateY(0deg);
    box-shadow: 
      0 0 30px rgba(255, 255, 255, 0.6),
      inset 0 0 20px rgba(255, 255, 255, 0.1);
  }
  50% {
    transform: scale(1.05) rotateY(10deg);
    box-shadow: 
      0 0 40px rgba(180, 180, 180, 0.8),
      inset 0 0 30px rgba(255, 255, 255, 0.2);
  }
}

@keyframes essence-pulse {
  0%, 100% {
    transform: scale(1);
    opacity: 1;
  }
  50% {
    transform: scale(1.3);
    opacity: 0.8;
  }
}

@keyframes layer-rotate {
  0% {
    transform: rotate(0deg);
  }
  100% {
    transform: rotate(360deg);
  }
}

@keyframes perception-twinkle {
  0%, 100% {
    opacity: 0.6;
    transform: scale(1);
  }
  50% {
    opacity: 1;
    transform: scale(1.2);
  }
}

@keyframes consciousness-expand {
  0% {
    transform: scale(0.8);
    opacity: 0.8;
  }
  70% {
    transform: scale(1.2);
    opacity: 0.3;
  }
  100% {
    transform: scale(1.4);
    opacity: 0;
  }
}
</style>