import * as THREE from './../../lib/three.js/three.module.js';

export default (object, domElement, props) => {
    const self = {
        object, domElement: domElement || window.document,
        moveSpeed: 1.0, lookSpeed: 0.005, autoForward: false,
        clock: new THREE.Clock(), enabled: true,
    };
    props && Object.keys(props).forEach(name => {
        if (!self.hasOwnProperty(name)) {
            console.warn(`ignored unsupport prop: ${name}.`);
            return;
        }
        self[name] = props[name];
    });

    const flags = {
        lookXValue: 0, lookYValue: 0, lookX: false, lookY: false, pointerLocked: false,
        holdA: false, holdB: false, holdX: false, holdY: false,
        forward: false, backward: false, left: false, right: false, up: false, down: false,
    };
    const actions = {
        lookX(value) { Number.isInteger(value) ? (flags.lookXValue += value) : (flags.lookX = !!value); },
        lookY(value) { Number.isInteger(value) ? (flags.lookYValue += value) : (flags.lookY = !!value); },
        A(active) { flags.holdA = !!active; },
        B(active) { flags.holdB = !!active; },
        X(active) { flags.holdX = !!active; },
        Y(active) { flags.holdY = !!active; },
        forward(active) { flags.forward = !!active; },
        backward(active) { flags.backward = !!active; },
        left(active) { flags.left = !!active; },
        right(active) { flags.right = !!active; },
        up(active) { flags.up = !!active; },
        down(active) { flags.down = !!active; },
    };
    self.keyMap = {};
    self.bindKey = (...binds) => binds.forEach(({ code, name, action }) => {
        self.keyMap[code] = actions[action];
        actions[action].key = name;
    });
    self.bindKey(
        { code: 'clickLeft', name: 'ClickLeft', action: 'A', },
        { code: 'clickRight', name: 'ClickRight', action: 'B', },
        { code: 87, name: 'KeyW', action: 'forward', },
        { code: 83, name: 'KeyS', action: 'backward', },
        { code: 65, name: 'KeyA', action: 'left', },
        { code: 68, name: 'KeyD', action: 'right', },
        { code: 32, name: 'Space', action: 'up', },
        { code: 17, name: 'ControlLeft', action: 'down', },
    );

    const onMouseDown = event => {
        self.domElement !== document && self.domElement.focus();
        if (!flags.pointerLocked) {
            self.domElement.requestPointerLock();
            return;
        }

        const action = self.keyMap[['clickLeft', null, 'clickRight'][event.button]];
        if (!action) {
            return;
        }
        action(true);
        event.preventDefault();
        event.stopPropagation();
    };
    const onMouseUp = event => {
        const action = self.keyMap[['clickLeft', null, 'clickRight'][event.button]];
        if (!action) {
            return;
        }
        action(false);
        event.preventDefault();
        event.stopPropagation();
    };
    const onMouseMove = event => {
        actions.lookX(event.movementX);
        actions.lookY(event.movementY);
    };
    document.addEventListener('pointerlockchange', () => {
        if (document.pointerLockElement === self.domElement) {
            flags.pointerLocked = true;
            self.domElement.addEventListener('mousemove', onMouseMove, false)
            return;
        }
        flags.pointerLocked = false;
        self.domElement.removeEventListener('mousemove', onMouseMove, false);
    }, false);
    const onKeyDown = event => {
        const action = self.keyMap[event.keyCode];
        if (!action) {
            return;
        }
        action(true);
        event.preventDefault();
        event.stopPropagation();
    };
    const onKeyUp = event => {
        const action = self.keyMap[event.keyCode];
        if (!action) {
            return;
        }
        action(false);
        event.preventDefault();
        event.stopPropagation();
    };

    const contextmenu = event => event.preventDefault();
    self.unmount = () => {
        self.domElement.removeEventListener('contextmenu', contextmenu, false);
        self.domElement.removeEventListener('mousedown', onMouseDown, false);
        self.domElement.removeEventListener('mouseup', onMouseUp, false);

        window.removeEventListener('keydown', onKeyDown, false);
        window.removeEventListener('keyup', onKeyUp, false);
    };
    self.mount = () => {
        self.domElement.addEventListener('contextmenu', contextmenu, false);
        self.domElement.addEventListener('mousedown', onMouseDown, false);
        self.domElement.addEventListener('mouseup', onMouseUp, false);

        window.addEventListener('keydown', onKeyDown, false);
        window.addEventListener('keyup', onKeyUp, false);
    }
    self.mount();

    const targetPosition = new THREE.Vector3(0, 0, 0);
    let offsetX = 0, offsetY = 0;
    self.update = delta => {
        if (!self.enabled) {
            return;
        }
        if (delta == null) {
            delta = self.clock.getDelta();
        }

        const actualMoveSpeed = delta * self.moveSpeed;
        (flags.forward || (self.autoForward && !flags.backward)) && self.object.translateZ(-(actualMoveSpeed));
        flags.backward && self.object.translateZ(actualMoveSpeed);
        flags.left && self.object.translateX(- actualMoveSpeed);
        flags.right && self.object.translateX(actualMoveSpeed);
        flags.up && self.object.translateY(actualMoveSpeed);
        flags.down && self.object.translateY(- actualMoveSpeed);

        const actualLookSpeed = delta * self.lookSpeed;
        flags.lookX && (flags.lookXValue = 100 * actualLookSpeed);
        flags.lookY && (flags.lookYValue = 100 * actualLookSpeed);
        if (flags.lookYValue == 0 && flags.lookXValue == 0) {
            return;
        }
        offsetX += flags.lookXValue * actualLookSpeed;
        flags.lookXValue = 0;
        offsetY -= flags.lookYValue * actualLookSpeed;
        flags.lookYValue = 0;
        offsetY = Math.max(-85, Math.min(85, offsetY))
        const theta = THREE.Math.degToRad(offsetX), phi = THREE.Math.degToRad(90 - offsetY)
        targetPosition.x = self.object.position.x + 100 * Math.sin(phi) * Math.cos(theta);
        targetPosition.y = self.object.position.y + 100 * Math.cos(phi);
        targetPosition.z = self.object.position.z + 100 * Math.sin(phi) * Math.sin(theta);
        self.object.lookAt(targetPosition);
    };

    return self;
};
