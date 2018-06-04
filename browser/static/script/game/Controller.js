import * as THREE from './../../lib/three.js/three.module.js';

import * as debug from './../debug.js';

/**
 * debounce, used such as combo attack.
 * @param {(times: number) => number} action times: trigger times before. return: timeout handler used to cancel.
 * @param {number} delay .
 */
export const debounce = (action, delay) => {
    let timeout;
    return (...args) => {
        clearTimeout(timeout);
        timeout = setTimeout(() => action(...args), delay);
        return timeout;
    };
};
/**
 * throttle, used such as magic cooldown.
 * @param {() => number} action return: timeout handler used to cancel.
 * @param {number} period .
 */
export const throttle = (action, period) => {
    let maybeCall;
    const call = (...args) => {
        action(...args);
        maybeCall = () => { };
        return setTimeout(() => maybeCall = call, period);
    };
    maybeCall = call;
    return (...args) => maybeCall(...args);
};

export const FirstPersonLogic = self => {
    const targetPosition = new THREE.Vector3(0, 0, 0);
    let autoMove = (period => {
        let autoMove = false;
        const trigger = throttle(() => autoMove = !autoMove, period);
        return () => { self.keys.R && trigger(); return autoMove; };
    })(300);
    let offsetX = 0, offsetY = 0;

    return delta => {
        if (!self.enabled) {
            return;
        }
        if (delta == null) {
            delta = self.getDelta();
        }

        const moveAxes = self.keys.LA;
        const actualMoveSpeed = delta * self.moveSpeed * (self.keys.LT ? 2 : 1) * (self.keys.L ? 0.5 : 1)
        self.object.translateX(moveAxes[0] * actualMoveSpeed);
        self.object.translateZ((
            Math.round(moveAxes[1]) != 0 ? moveAxes[1] : (autoMove() ? -1 : 0)
        ) * actualMoveSpeed);

        self.keys.up && self.object.translateY(delta * self.moveSpeed);
        self.keys.down && self.object.translateY(- delta * self.moveSpeed);

        const lookAxes = self.keys.RA;
        offsetX += lookAxes[0] * delta * self.lookSpeed;
        offsetY -= lookAxes[1] * delta * self.pitchSpeed;
        self.keymouse.mouseAxes = [0, 0]; // 鼠标移动不自动回原点，需要清零
        offsetY = Math.max(-85, Math.min(85, offsetY))
        const theta = THREE.Math.degToRad(offsetX), phi = THREE.Math.degToRad(90 - offsetY)
        targetPosition.x = self.object.position.x + 100 * Math.sin(phi) * Math.cos(theta);
        targetPosition.y = self.object.position.y + 100 * Math.cos(phi);
        targetPosition.z = self.object.position.z + 100 * Math.sin(phi) * Math.sin(theta);
        self.object.lookAt(targetPosition);
    };
}

export default (object, domElement, props) => {
    const self = {
        object, domElement: domElement || window.document,
        clock: null, logic: null, enabled: true,
        moveSpeed: 5.0, jumpSpeed: 5.0, lookSpeed: 1.0, pitchSpeed: 0.5,
    };
    props && Object.keys(props).forEach(name => {
        if (!self.hasOwnProperty(name)) {
            console.warn(`ignored unsupport prop: ${name}.`);
            return;
        }
        self[name] = props[name];
    });

    // 手柄标识
    const gamepad = {
        mapper: {
            LA: [0, 0], RA: [0, 0],
            LB: false, RB: false, // 肩键
            LT: false, RT: false, // 肩背键
            L: false, R: false, // 摇杆按下
            up: false, down: false, left: false, right: false, // 左十字键
            A: false, B: false, X: false, Y: false, // 右四键
            select: false, start: false, // 中间两键
        },
        get leftAxes() {
            const gamepad = this._gamepad;
            return gamepad ? gamepad.axes.slice(0, 2) : [0, 0];
        },
        get rightAxes() {
            const gamepad = this._gamepad;
            return gamepad ? gamepad.axes.slice(2, 4) : [0, 0];
        },
        of(key) {
            const code = this.mapper[key].code
            if (this.hasOwnProperty(code)) {
                return this[code];
            }
            const gamepad = this._gamepad;
            return gamepad && gamepad.buttons[code].pressed;
        },

        get _gamepad() { return navigator.getGamepads()[0]; },
    };
    self.gamepad = gamepad;
    // 键鼠标识
    const keymouse = {
        mapper: {
            mainUp: false, mainDown: false, mainLeft: false, mainRight: false,
            ...gamepad.mapper,
        },
        get mainAxes() {
            const x = (this.of('mainLeft') ? -1 : 0) + (this.of('mainRight') ? 1 : 0);
            const y = (this.of('mainUp') ? -1 : 0) + (this.of('mainDown') ? 1 : 0);
            const d = Math.sqrt(x ** 2 + y ** 2);
            return d == 0 ? null : [x / d, y / d];
        },
        get mouseAxes() { return this._mouseAxes; }, set mouseAxes([x, y]) {
            if (x == 0 && y == 0) {
                this._mouseAxes = null;
                return;
            }
            this._mouseAxes = [x / 100, y / 100];
        },
        of(key) {
            const code = this.mapper[key].code;
            return this[code] || this.buttons[code];
        },

        buttons: { clickLeft: false, clickRight: false, /* keyCode: pressed */ },
        pointerLocked: false,
    };
    self.keymouse = keymouse;
    // 绑定键位
    const bindKeys = (self, ...binds) => binds.forEach(
        ([key, code, name]) => self.mapper[key] = { code, name: name || key }
    );
    self.bindGamepad = bindKeys.bind(null, gamepad);
    self.bindKeymouse = bindKeys.bind(null, keymouse);
    self.bindGamepad(
        ['LA', 'leftAxes'], ['RA', 'rightAxes'],
        ['LB', 4], ['RB', 5],
        ['LT', 6], ['RT', 7],
        ['L', 10], ['R', 11],
        ['up', 12], ['down', 13], ['left', 14], ['right', 15],
        ['A', 0], ['B', 1], ['X', 2], ['Y', 3],
        ['select', 8], ['start', 9],
    );
    self.bindKeymouse(
        ['mainUp', 87, 'KeyW'], ['mainDown', 83, 'KeyS'], ['mainLeft', 65, 'KeyA'], ['mainRight', 68, 'KeyD'],
        ['LA', 'mainAxes'], ['RA', 'mouseAxes'],
        ['LB', 'clickLeft', 'ClickLeft'], ['RB', 'clickRight', 'ClickRight'],
        ['LT', 17, 'ControlLeft'], ['RT', 32, 'Space'],
        ['L', 16, 'ShiftLeft'], ['R', 20, 'CapsLock'],
        ['up', 38, 'ArrowUp'], ['down', 40, 'ArrowDown'], ['left', 37, 'ArrowLeft'], ['right', 39, 'ArrowRight'],
        ['A', /* unused */], ['B', /* unused */], ['X', 81, 'KeyQ'], ['Y', 69, 'KeyE'],
        ['select', 27, 'Escape'], ['start', 13, 'Enter'],
    );
    // 键位状态
    self.keys = { ...gamepad.mapper };
    Object.defineProperties(self.keys, Object.keys(self.keys).map(key => [
        key, () => keymouse.of(key) || gamepad.of(key),
    ]).reduce((props, [key, get]) => ({ ...props, [key]: { get } }), {}));

    // 鼠标事件
    const onMouseDown = event => {
        self.domElement !== document && self.domElement.focus();
        if (!keymouse.pointerLocked) {
            self.domElement.requestPointerLock();
            return;
        }

        event.preventDefault();
        event.stopPropagation();
        keymouse.buttons[['clickLeft', null, 'clickRight'][event.button]] = true;
    };
    const onMouseUp = event => {
        event.preventDefault();
        event.stopPropagation();
        keymouse.buttons[['clickLeft', null, 'clickRight'][event.button]] = false;
    };
    const onMouseMove = event => keymouse.mouseAxes = [event.movementX, event.movementY];
    document.addEventListener('pointerlockchange', () => {
        if (document.pointerLockElement === self.domElement) {
            keymouse.pointerLocked = true;
            self.domElement.addEventListener('mousemove', onMouseMove, false)
            return;
        }
        keymouse.pointerLocked = false;
        self.domElement.removeEventListener('mousemove', onMouseMove, false);
    }, false);
    self.freePointerLock = () => document.exitPointerLock();

    // 键盘事件
    const onKeyDown = event => {
        event.preventDefault();
        event.stopPropagation();
        keymouse.buttons[event.keyCode] = true;
    };
    const onKeyUp = event => {
        event.preventDefault();
        event.stopPropagation();
        keymouse.buttons[event.keyCode] = false;
    };

    // 手柄事件
    const onGamepadConnect = e => {
        console.log("Gamepad connected at index %d: %s. %d buttons, %d axes.",
            e.gamepad.index, e.gamepad.id,
            e.gamepad.buttons.length, e.gamepad.axes.length,
        );
    };
    const onGamepadDisconnect = e => {
        // if (gamepad == e.gamepad) {
        //     gamepad = null;
        // }
        console.log("Gamepad disconnected from index %d: %s", e.gamepad.index, e.gamepad.id);
    };

    // 注册控制事件
    const contextmenu = event => event.preventDefault();
    const mountHelp = mount => {
        const action = `${mount ? 'add' : 'remove'}EventListener`;

        self.domElement[action]('contextmenu', contextmenu, false);
        self.domElement[action]('mousedown', onMouseDown, false);
        self.domElement[action]('mouseup', onMouseUp, false);

        window[action]('keydown', onKeyDown, false);
        window[action]('keyup', onKeyUp, false);
        window[action]('gamepadconnected', onGamepadConnect, false);
        window[action]('gamepaddisconnected', onGamepadDisconnect, false);
    };
    self.unmount = mountHelp.bind(this, false);
    self.mount = mountHelp.bind(this, true);
    self.mount();

    // Mixin
    self.clock = self.clock || new THREE.Clock();
    self.getDelta = () => self.clock.getDelta();
    self.logic = self.logic || FirstPersonLogic(self);
    self.update = delta => self.logic(delta);

    debug.exports('controller', exports => Object.defineProperties(exports, {
        default: { get: () => self },
    }));

    return self;
};
